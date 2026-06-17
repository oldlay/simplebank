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

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid-Email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "ok",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{
						String: newName,
						Valid:  true,
					},
					Email: sql.NullString{
						String: newEmail,
						Valid:  true,
					},
				}
				updateUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updateUser, nil)

			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, updatedUser.Username, user.Username)
				require.Equal(t, updatedUser.Email, newEmail)
				require.Equal(t, updatedUser.FullName, newName)

			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorizationToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &invalidEmail,
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
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

			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
