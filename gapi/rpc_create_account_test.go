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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func randomAccount(username string) (Account db.Account) {
	Account = db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    username,
		Balance:  decimal.Zero,
		Currency: util.RandomCurrency(),
	}

	return
}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		req           *pb.CreateAccountRequest
		buildContext  func(t *testing.T, tokenMaker *token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateAccountResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.CreateAccountRequest{
				Currency: account.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  account.Balance,
					Currency: account.Currency,
				}

				createAccount := db.Account{
					ID:        account.ID,
					Owner:     account.Owner,
					Balance:   account.Balance,
					Currency:  account.Currency,
					CreatedAt: account.CreatedAt,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(createAccount, nil)

			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdAccount := res.GetAccount()
				require.Equal(t, createdAccount.Owner, account.Owner)
				require.Equal(t, createdAccount.Currency, account.Currency)
				require.Equal(t, createdAccount.Amount, util.ShopDecimalToProto(account.Balance))
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateAccountRequest{
				Currency: account.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InvalidCurrency",
			req: &pb.CreateAccountRequest{
				Currency: "CNY",
			},
			buildContext: func(t *testing.T, tokenMaker *token.Maker) context.Context {
				return newContextWithBearerToken(t, *tokenMaker, account.Owner, time.Minute, util.DepositorRole, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
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

			res, err := server.CreateAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
