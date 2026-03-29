package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	identityV1 "go-wind-admin/api/gen/go/identity/service/v1"
	"go-wind-admin/app/game/service/internal/data"
)

type UserProxyService struct {
	log  *log.Helper
	data *data.AdminClient
}

func NewUserProxyService(
	ctx *bootstrap.Context,
	adminClient *data.AdminClient,
) *UserProxyService {
	return &UserProxyService{
		log:  ctx.NewLoggerHelper("game/service/user-proxy"),
		data: adminClient,
	}
}

func (s *UserProxyService) GetUserById(ctx context.Context, id uint32) (*identityV1.User, error) {
	s.log.Debugf("GetUserById called with id: %d", id)
	return s.data.UserService.Get(ctx, &identityV1.GetUserRequest{
		QueryBy: &identityV1.GetUserRequest_Id{Id: id},
	})
}
