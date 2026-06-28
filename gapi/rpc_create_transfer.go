package gapi

import (
	"context"
	"errors"
	"fmt"

	db "github.com/oldlay/simplebank/db/sqlc"
	"github.com/oldlay/simplebank/pb"
	"github.com/oldlay/simplebank/util"
	"github.com/oldlay/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	//fromAccount violation check
	violations := validateCreateTransferRequest(req)

	fromAccount, err := ValidateAccount(ctx, server, req.FromAccountId, req.Currency)
	if err != nil {
		fmt.Println("ValidateAccount")
		violations = append(violations, fieldViolation("From_account_invalid", err))
	}

	//toAccount violation check
	getAccountArg := db.GetAccountFromOwnerParams{
		Owner:    req.ToOwner,
		Currency: req.ToCurrency,
	}
	toAccountId, err := server.store.GetAccountFromOwner(ctx, getAccountArg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get account: %s", err)
	}

	_, err = ValidateAccount(ctx, server, toAccountId, req.Currency)
	if err != nil {
		fmt.Println("ValidateAccount")
		violations = append(violations, fieldViolation("To_account_invalid", err))
	}

	if violations != nil {
		fmt.Println("validateCreateTransferRequest")
		return nil, invalidArgumentError(violations)
	}

	if fromAccount.Owner != authPayload.Username {
		return nil, status.Error(codes.PermissionDenied, "account doesn't belong to you")
	}

	amount, err := util.ProtoToShopDecimal(req.Amount)
	if err != nil {
		fmt.Println("rotoToShopDecimal")
		return nil, invalidArgumentError(violations)
	}

	//invalid amount
	if fromAccount.Balance.LessThan(amount.Abs()) {
		fmt.Println("LessThan")
		return nil, invalidArgumentError(violations)
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountId,
		ToAccountID:   toAccountId,
		Amount:        amount,
	}
	transfer, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transfer: %s", err)
	}
	rsp := &pb.CreateTransferResponse{
		Transfer:    convertTransfer(transfer.Transfer),
		FromAccount: convertAccount(transfer.FromAccount),
		ToAccount:   convertAccount(transfer.ToAccount),
		FromEntry:   convertEntry(transfer.FromEntry),
		ToEntry:     convertEntry(transfer.ToEntry),
	}
	return rsp, nil
}

func validateCreateTransferRequest(req *pb.CreateTransferRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateId(req.GetFromAccountId()); err != nil {
		violations = append(violations, fieldViolation("Transfer_from_id", err))
	}
	if err := val.ValidateCurrency(req.GetCurrency()); err != nil {
		violations = append(violations, fieldViolation("Transfer_currency", err))
	}
	if err := val.ValidateAmount(req.GetAmount()); err != nil {
		violations = append(violations, fieldViolation("Transfer_currency", err))
	}
	if req.GetToOwner() == "" {
		violations = append(violations, fieldViolation("to_owner", fmt.Errorf("required")))
	}
	if err := val.ValidateCurrency(req.GetToCurrency()); err != nil {
		violations = append(violations, fieldViolation("to_currency", err))
	}

	return violations
}

func ValidateAccount(ctx context.Context, server *Server, accountID int64, currency string) (db.Account, error) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return account, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return account, status.Errorf(codes.Internal, "failed to get account: %s", err)
	}

	if account.Currency != currency {
		return account, status.Errorf(codes.InvalidArgument, "account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
	}
	return account, nil
}
