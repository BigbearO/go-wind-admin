package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-admin/app/game/service/internal/data/ent"
	"go-wind-admin/app/game/service/internal/data/ent/predicate"

	gameV1 "go-wind-admin/api/gen/go/game/service/v1"
)

type GameAccountRepo interface {
	List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListGameAccountResponse, error)
	Get(ctx context.Context, req *gameV1.GetGameAccountRequest) (*gameV1.GameAccount, error)
	Create(ctx context.Context, req *gameV1.CreateGameAccountRequest) (*gameV1.GameAccount, error)
	Update(ctx context.Context, req *gameV1.UpdateGameAccountRequest) error
	Delete(ctx context.Context, req *gameV1.DeleteGameAccountRequest) error
	Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error)
}

type gameAccountRepo struct {
	log        *log.Helper
	entClient  *entgo.EntClient[*ent.Client]
	mapper     *mapper.CopierMapper[gameV1.GameAccount, ent.GameAccount]
	repository *entgo.Repository[
		ent.GameAccountQuery, ent.GameAccountSelect,
		ent.GameAccountCreate, ent.GameAccountCreateBulk,
		ent.GameAccountUpdate, ent.GameAccountUpdateOne,
		ent.GameAccountDelete,
		predicate.GameAccount,
		gameV1.GameAccount, ent.GameAccount,
	]
}

func NewGameAccountRepo(
	ctx *bootstrap.Context,
	entClient *entgo.EntClient[*ent.Client],
) GameAccountRepo {
	repo := &gameAccountRepo{
		log:       ctx.NewLoggerHelper("game/repo/game-account"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[gameV1.GameAccount, ent.GameAccount](),
	}

	repo.init()

	return repo
}

func (r *gameAccountRepo) init() {
	r.repository = entgo.NewRepository[
		ent.GameAccountQuery, ent.GameAccountSelect,
		ent.GameAccountCreate, ent.GameAccountCreateBulk,
		ent.GameAccountUpdate, ent.GameAccountUpdateOne,
		ent.GameAccountDelete,
		predicate.GameAccount,
		gameV1.GameAccount, ent.GameAccount,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *gameAccountRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().GameAccount.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, err
	}

	return count, nil
}

func (r *gameAccountRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListGameAccountResponse, error) {
	builder := r.entClient.Client().GameAccount.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &gameV1.ListGameAccountResponse{Total: 0, Items: nil}, nil
	}

	return &gameV1.ListGameAccountResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *gameAccountRepo) Get(ctx context.Context, req *gameV1.GetGameAccountRequest) (*gameV1.GameAccount, error) {
	result, err := r.entClient.Client().GameAccount.Get(ctx, req.GetId())
	if err != nil {
		r.log.Errorf("query one failed: %s", err.Error())
		return nil, err
	}

	return r.mapper.ToDTO(result), nil
}

func (r *gameAccountRepo) Create(ctx context.Context, req *gameV1.CreateGameAccountRequest) (*gameV1.GameAccount, error) {
	if req.Data == nil {
		return nil, nil
	}

	builder := r.entClient.Client().GameAccount.Create()

	data := req.Data
	if data.UserAccount != nil {
		builder.SetUserAccount(*data.UserAccount)
	}
	if data.Credential != nil {
		builder.SetCredential(*data.Credential)
	}
	if data.StartTime != nil {
		builder.SetStartTime(data.StartTime.AsTime())
	}
	if data.ExpireTime != nil {
		builder.SetExpireTime(data.ExpireTime.AsTime())
	}
	if data.Type != nil {
		builder.SetType(*data.Type)
	}
	if data.Enabled != nil {
		builder.SetEnabled(*data.Enabled)
	}
	if data.EmulatorId != nil {
		builder.SetEmulatorID(*data.EmulatorId)
	}

	if data.TenantId != nil {
		builder.SetTenantID(*data.TenantId)
	}

	result, err := builder.Save(ctx)
	if err != nil {
		r.log.Errorf("create failed: %s", err.Error())
		return nil, err
	}

	return r.mapper.ToDTO(result), nil
}

func (r *gameAccountRepo) Update(ctx context.Context, req *gameV1.UpdateGameAccountRequest) error {
	if req.Data == nil {
		return nil
	}

	builder := r.entClient.Client().GameAccount.UpdateOneID(req.GetId())

	data := req.Data
	if data.UserAccount != nil {
		builder.SetUserAccount(*data.UserAccount)
	}
	if data.Credential != nil {
		builder.SetCredential(*data.Credential)
	}
	if data.StartTime != nil {
		builder.SetStartTime(data.StartTime.AsTime())
	}
	if data.ExpireTime != nil {
		builder.SetExpireTime(data.ExpireTime.AsTime())
	}
	if data.Type != nil {
		builder.SetType(*data.Type)
	}
	if data.Enabled != nil {
		builder.SetEnabled(*data.Enabled)
	}
	if data.EmulatorId != nil {
		builder.SetEmulatorID(*data.EmulatorId)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		r.log.Errorf("update failed: %s", err.Error())
		return err
	}

	return nil
}

func (r *gameAccountRepo) Delete(ctx context.Context, req *gameV1.DeleteGameAccountRequest) error {
	err := r.entClient.Client().GameAccount.DeleteOneID(req.GetId()).Exec(ctx)
	if err != nil {
		r.log.Errorf("delete failed: %s", err.Error())
		return err
	}
	return nil
}
