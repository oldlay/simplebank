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

func (server *Server) ListAccount(ctx context.Context, req *pb.ListAccountRequest) (*pb.ListAccountResponse, error) {
	violations := validateListAccountRequest(req)

	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.ListAccountParams{
		Owner:  authPayload.Username,
		Limit:  int32(req.PageSize),
		Offset: int32(req.PageId-1) * int32(req.PageSize),
	}

	accounts, err := server.store.ListAccount(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to List account: %s", err)
	}
	for _, at := range accounts {
		if at.Owner != authPayload.Username {
			return nil, status.Errorf(codes.Internal, "No authorization: %s", err)
		}
	}

	var pbAccount []*pb.Account
	for _, at := range accounts {
		pbAccount = append(pbAccount, convertAccount(at))
	}

	rsp := &pb.ListAccountResponse{
		Accounts: pbAccount,
	}
	return rsp, nil
}

func validateListAccountRequest(req *pb.ListAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateId(req.GetPageId()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}
	if err := val.ValidateId(req.GetPageSize()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}

	return violations
}
