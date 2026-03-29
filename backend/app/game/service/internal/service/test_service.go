package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	gameV1 "go-wind-admin/api/gen/go/game/service/v1"
	identityV1 "go-wind-admin/api/gen/go/identity/service/v1"
	"go-wind-admin/app/game/service/internal/data"
)

type TestService struct {
	log  *log.Helper
	data *data.AdminClient
}

func NewTestService(
	ctx *bootstrap.Context,
	adminClient *data.AdminClient,
) *TestService {
	return &TestService{
		log:  ctx.NewLoggerHelper("game/service/test"),
		data: adminClient,
	}
}

func (s *TestService) GetUser(ctx context.Context, req *gameV1.GetUserRequest) (*gameV1.GetUserResponse, error) {
	s.log.Debugf("GetUser called with id: %d", req.GetId())

	user, err := s.data.UserService.Get(ctx, &identityV1.GetUserRequest{
		QueryBy: &identityV1.GetUserRequest_Id{Id: req.GetId()},
	})
	if err != nil {
		s.log.Errorf("GetUser failed: %v", err)
		return nil, err
	}

	return &gameV1.GetUserResponse{
		Id:       user.GetId(),
		Username: user.GetUsername(),
		Nickname: user.GetNickname(),
		Email:    user.GetEmail(),
		Mobile:   user.GetMobile(),
		TenantId: user.GetTenantId(),
	}, nil
}
