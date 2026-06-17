package db

import "context"

// CreateUserTxParams contains the input parameters of the transfer transaction
type CreateUserTxParams struct {
	CreateUserParams
	//callback func, be excuted after the user is inserted,output will decide whether rollback
	AfterCreate func(user User) error
}

type CreateUserTxResult struct {
	User User
}

//var txKey = struct{}{}

// CreateUserTx performs a money transfer from one account to the other.
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
}
