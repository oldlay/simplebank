package gapi

import (
	db "github.com/oldlay/simplebank/db/sqlc"
	"github.com/oldlay/simplebank/pb"
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
	}
}
