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

func TestCreateTransferAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	user3, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)
	account3 := randomAccount(user3.Username)

	account1.Balance = decimal.NewFromInt(1000)
	account1.Currency = util.USD
	account2.Currency = util.USD
	account3.Currency = util.EUR

	transferAmount := util.ShopDecimalToProto(decimal.NewFromInt(10))

	transfer := db.Transfer{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        decimal.NewFromInt(10),
	}

	transferedAccount1 := db.Account{
		Owner:     account1.Owner,
		Balance:   account1.Balance.Sub(decimal.NewFromInt(10)),
		Currency:  account1.Currency,
		CreatedAt: account1.CreatedAt,
	}
	transferedAccount2 := db.Account{
		Owner:     account2.Owner,
		Balance:   account2.Balance.Add(decimal.NewFromInt(10)),
		Currency:  account2.Currency,
		CreatedAt: account2.CreatedAt,
	}
	entry1 := db.Entry{
		AccountID: account1.ID,
		Amount:    decimal.NewFromInt(-10),
	}
	entry2 := db.Entry{
		AccountID: account2.ID,
		Amount:    decimal.NewFromInt(10),
	}

	testCases := []struct {
		name          string
		req           *pb.CreateTransferRequest
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateTransferResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToOwner:       account2.Owner,
				ToCurrency:    account2.Currency,
				Amount:        transferAmount,
				Currency:      account1.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute, user1.Role, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				GetArg := db.GetAccountFromOwnerParams{
					Owner:    account2.Owner,
					Currency: account2.Currency,
				}
				CreatedTransfer := db.TransferTxResult{
					Transfer:    transfer,
					FromAccount: transferedAccount1,
					ToAccount:   transferedAccount2,
					FromEntry:   entry1,
					ToEntry:     entry2,
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
				store.EXPECT().GetAccountFromOwner(gomock.Any(), gomock.Eq(GetArg)).Times(1).Return(account2.ID, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(CreatedTransfer, nil)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedFromAccount := res.GetFromAccount()
				FromAccountAmount, _ := util.ProtoToShopDecimal(updatedFromAccount.Amount)
				require.True(t, FromAccountAmount.Add(decimal.NewFromInt(10)).Equal(account1.Balance))

				updatedToAccount := res.GetToAccount()
				ToAccountAmount, _ := util.ProtoToShopDecimal(updatedToAccount.Amount)
				require.True(t, ToAccountAmount.Sub(decimal.NewFromInt(10)).Equal(account2.Balance))
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToOwner:       account2.Owner,
				ToCurrency:    account2.Currency,
				Amount:        transferAmount,
				Currency:      account1.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute, user1.Role, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				GetArg := db.GetAccountFromOwnerParams{
					Owner:    account2.Owner,
					Currency: account2.Currency,
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
				store.EXPECT().GetAccountFromOwner(gomock.Any(), gomock.Eq(GetArg)).Times(1).Return(account2.ID, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "Unauthorization",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToOwner:       account2.Owner,
				ToCurrency:    account2.Currency,
				Amount:        transferAmount,
				Currency:      account1.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidCurrency",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToOwner:       account3.Owner,
				ToCurrency:    account3.Currency,
				Amount:        transferAmount,
				Currency:      account3.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute, user1.Role, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account3.ID,
					Amount:        amount,
				}
				GetArg := db.GetAccountFromOwnerParams{
					Owner:    account3.Owner,
					Currency: account3.Currency,
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
				store.EXPECT().GetAccountFromOwner(gomock.Any(), gomock.Eq(GetArg)).Times(1).Return(account3.ID, nil)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidAmount",
			req: &pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToOwner:       account2.Owner,
				ToCurrency:    account2.Currency,
				Amount:        util.ShopDecimalToProto(decimal.NewFromInt(1000000)),
				Currency:      account2.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute, user1.Role, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				GetArg := db.GetAccountFromOwnerParams{
					Owner:    account2.Owner,
					Currency: account2.Currency,
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)
				store.EXPECT().GetAccountFromOwner(gomock.Any(), gomock.Eq(GetArg)).Times(1).Return(account2.ID, nil)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidId",
			req: &pb.CreateTransferRequest{
				FromAccountId: 0,
				ToOwner:       "qingqing",
				ToCurrency:    "CNY",
				Amount:        transferAmount,
				Currency:      account2.Currency,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user1.Username, time.Minute, user1.Role, token.TokenTypeAccessToken)
			},
			buildStubs: func(store *mockdb.MockStore) {
				amount, _ := util.ProtoToShopDecimal(transferAmount)
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				GetArg := db.GetAccountFromOwnerParams{
					Owner:    "qingqing",
					Currency: "CNY",
				}

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetAccountFromOwner(gomock.Any(), gomock.Eq(GetArg)).Times(1).Return(int64(0), sql.ErrNoRows)

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
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

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.CreateTransfer(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
