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

func (server *Server) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	if authPayload.Role != util.BankerRole {
		return nil, status.Error(codes.PermissionDenied, "cannot update other account's info")
	}

	balance, err := util.ProtoToShopDecimal(req.GetBalance())
	if err != nil {
		return nil, decimalTransError(err)
	}

	arg := db.UpdateAccountParams{
		ID:      req.GetId(),
		Balance: balance,
	}

	account, err := server.store.UpdateAccount(ctx, arg)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}
	if authPayload.Username != account.Owner {
		return nil, unauthenticatedError(err)
	}

	rsp := &pb.UpdateAccountResponse{
		Account: convertAccount(account),
	}
	return rsp, nil
}

func validateUpdateAccountRequest(req *pb.UpdateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateId(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}

	return violations
}
