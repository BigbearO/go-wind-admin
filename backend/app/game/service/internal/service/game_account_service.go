package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/game/service/internal/data"

	gameV1 "go-wind-admin/api/gen/go/game/service/v1"
)

type GameAccountService struct {
	gameV1.UnimplementedGameAccountServiceServer

	log *log.Helper

	gameAccountRepo data.GameAccountRepo
}

func NewGameAccountService(
	ctx *bootstrap.Context,
	gameAccountRepo data.GameAccountRepo,
) *GameAccountService {
	return &GameAccountService{
		log:             ctx.NewLoggerHelper("game/service/game-account-service"),
		gameAccountRepo: gameAccountRepo,
	}
}

func (s *GameAccountService) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListGameAccountResponse, error) {
	return s.gameAccountRepo.List(ctx, req)
}

func (s *GameAccountService) Get(ctx context.Context, req *gameV1.GetGameAccountRequest) (*gameV1.GameAccount, error) {
	return s.gameAccountRepo.Get(ctx, req)
}

func (s *GameAccountService) Create(ctx context.Context, req *gameV1.CreateGameAccountRequest) (*emptypb.Empty, error) {
	_, err := s.gameAccountRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GameAccountService) Update(ctx context.Context, req *gameV1.UpdateGameAccountRequest) (*emptypb.Empty, error) {
	err := s.gameAccountRepo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GameAccountService) Delete(ctx context.Context, req *gameV1.DeleteGameAccountRequest) (*emptypb.Empty, error) {
	err := s.gameAccountRepo.Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
