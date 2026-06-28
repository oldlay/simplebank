package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/oldlay/simplebank/db/mock"
	db "github.com/oldlay/simplebank/db/sqlc"
	"github.com/oldlay/simplebank/pb"
	"github.com/oldlay/simplebank/token"
	"github.com/oldlay/simplebank/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		req           *pb.GetAccountRequest
		buildContext  func(t *testing.T, tokenMaker *token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.GetAccountResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := account.ID
				getAccount := db.Account{
					ID:        account.ID,
					Owner:     account.Owner,
					Balance:   account.Balance,
					Currency:  account.Currency,
					CreatedAt: account.CreatedAt,
				}
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(getAccount, nil)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				getedAccount := res.GetAccount()
				require.Equal(t, getedAccount.Owner, account.Owner)
				require.Equal(t, getedAccount.Currency, account.Currency)
				require.Equal(t, getedAccount.Amount, util.ShopDecimalToProto(account.Balance))
			},
		},
		{
			name: "InternalError",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())

			},
		},
		{
			name: "Unauthorization",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return context.Background()
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidId",
			req: &pb.GetAccountRequest{
				Id: 0,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())

			},
		},
		{
			name: "IdNotFound",
			req: &pb.GetAccountRequest{
				Id: 1001,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			storectrl := gomock.NewController(t)
			defer storectrl.Finish()

			store := mockdb.NewMockStore(storectrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)

			ctx := tc.buildContext(t, &server.tokenMaker)

			res, err := server.GetAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
