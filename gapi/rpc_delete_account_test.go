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

func TestDeleteAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		req           *pb.DeleteAccountRequest
		buildContext  func(t *testing.T, tokenMaker *token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.DeleteAccountResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.DeleteAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.BankerRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, res.GetDeleteResult(), "successful delete a account!")
			},
		},
		{
			name: "Unauthorization",
			req: &pb.DeleteAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return context.Background()
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())

			},
		},
		{
			name: "InvalidRequest",
			req: &pb.DeleteAccountRequest{
				Id: -1,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.BankerRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "IdNotFound",
			req: &pb.DeleteAccountRequest{
				Id: 100001,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.BankerRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(1).Return(db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.DeleteAccountRequest{
				Id: account.ID,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.BankerRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
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

			res, err := server.DeleteAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
