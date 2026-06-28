package gapi

import (
	"context"
	"errors"

	db "github.com/oldlay/simplebank/db/sqlc"
	"github.com/oldlay/simplebank/pb"
	"github.com/oldlay/simplebank/util"
	"github.com/oldlay/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	authoPayload, err := server.authorizeUser(ctx, []string{util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateDeleteAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	if authoPayload.Role != util.BankerRole {
		return nil, status.Error(codes.PermissionDenied, "cannot delete user's info")
	}

	err = server.store.DeleteAccount(ctx, req.GetId())

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to Delete user: %s", err)
	}

	rsp := &pb.DeleteAccountResponse{
		DeleteResult: "successful delete a account!",
	}
	return rsp, nil
}

func validateDeleteAccountRequest(req *pb.DeleteAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateId(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}

	return violations
}
