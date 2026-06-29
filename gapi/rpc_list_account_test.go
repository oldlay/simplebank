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

func TestListAccount(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	accounts := make([]db.Account, n)

	for i := range n {
		accounts[i] = randomAccount(user.Username)
	}

	testCases := []struct {
		name          string
		req           *pb.ListAccountRequest
		buildContext  func(t *testing.T, tokenMaker *token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.ListAccountResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.ListAccountRequest{
				PageId:   1,
				PageSize: int64(n),
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, accounts[0].Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				listedAccounts := res.GetAccounts()
				for i := 0; i < n; i++ {
					require.Equal(t, listedAccounts[0].Owner, accounts[0].Owner)
					require.Equal(t, listedAccounts[0].Currency, accounts[0].Currency)
					require.Equal(t, listedAccounts[0].Amount, util.ShopDecimalToProto(accounts[0].Balance))
				}

			},
		},
		{
			name: "Unauthorization",
			req: &pb.ListAccountRequest{
				PageId:   1,
				PageSize: int64(n),
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return context.Background()
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())

			},
		},
		{
			name: "InvalidRequest",
			req: &pb.ListAccountRequest{
				PageId:   -1,
				PageSize: int64(n),
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, accounts[0].Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "IdNotFound",
			req: &pb.ListAccountRequest{
				PageId:   10,
				PageSize: int64(n),
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, accounts[0].Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.ListAccountRequest{
				PageId:   1,
				PageSize: int64(n),
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, accounts[0].Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
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

			res, err := server.ListAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
