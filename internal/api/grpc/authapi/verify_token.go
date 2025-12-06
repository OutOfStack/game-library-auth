package authapi

import (
	"context"
	"strings"

	pb "github.com/OutOfStack/game-library-auth/pkg/proto/authapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VerifyToken verifies JWT access token
func (s *AuthService) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	_, span := tracer.Start(ctx, "VerifyToken")
	defer span.End()

	token := req.GetToken()
	if strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.InvalidArgument, "empty token")
	}

	valid := s.authFacade.ValidateAccessToken(token)

	resp := &pb.VerifyTokenResponse{}
	resp.SetValid(valid)
	return resp, nil
}
