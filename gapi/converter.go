package gapi

import (
	db "github.com/oldlay/simplebank/db/sqlc"
	"github.com/oldlay/simplebank/pb"
	"github.com/oldlay/simplebank/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertUser converts a db.User to a pb.User,
// because we don't want to expose the hashed password field in the pb.User struct.
// So it wise to separate db model and pb model api.
func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		Role:              user.Role,
	}
}
func convertAccount(account db.Account) *pb.Account {
	return &pb.Account{
		Id:       account.ID,
		Owner:    account.Owner,
		Currency: account.Currency,
		Amount:   util.ShopDecimalToProto(account.Balance),
		CreateAt: timestamppb.New(account.CreatedAt),
	}
}

func convertTransfer(Transfer db.Transfer) *pb.Transfer {
	return &pb.Transfer{
		FromAccountId: Transfer.FromAccountID,
		ToAccountId:   Transfer.ToAccountID,
		Amount:        util.ShopDecimalToProto(Transfer.Amount),
		CreatedAt:     timestamppb.New(Transfer.CreatedAt),
	}

}

func convertEntry(Entry db.Entry) *pb.Entry {
	return &pb.Entry{
		Id:        Entry.ID,
		AccountId: Entry.AccountID,
		Amount:    util.ShopDecimalToProto(Entry.Amount),
	}
}
