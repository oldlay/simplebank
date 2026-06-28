package gapi

import (
	"context"

	"github.com/oldlay/simplebank/pb"
	"github.com/oldlay/simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetExchangeRate(ctx context.Context, req *pb.GetExchangeRateRequest) (*pb.GetExchangeRateResponse, error) {
	result, err := util.GetExchangeAPI()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch exchange rates: %s", err)
	}

	return &pb.GetExchangeRateResponse{
		Result: result.Result,
		Date:   result.Date,
		Rates:  result.Rate,
	}, nil
}
