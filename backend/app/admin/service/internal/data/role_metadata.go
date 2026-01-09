package data

import (
	"context"
	"go-wind-admin/app/admin/service/internal/data/ent/rolemetadata"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"

	userV1 "go-wind-admin/api/gen/go/user/service/v1"
)

type RoleMetadataRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[userV1.RoleMetadata, ent.RoleMetadata]

	repository *entCrud.Repository[
		ent.RoleMetadataQuery, ent.RoleMetadataSelect,
		ent.RoleMetadataCreate, ent.RoleMetadataCreateBulk,
		ent.RoleMetadataUpdate, ent.RoleMetadataUpdateOne,
		ent.RoleMetadataDelete,
		predicate.RoleMetadata,
		userV1.RoleMetadata, ent.RoleMetadata,
	]
}

func NewRoleMetadataRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *RoleMetadataRepo {
	repo := &RoleMetadataRepo{
		log:       ctx.NewLoggerHelper("role-metadata/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[userV1.RoleMetadata, ent.RoleMetadata](),
	}

	repo.init()

	return repo
}

func (r *RoleMetadataRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.RoleMetadataQuery, ent.RoleMetadataSelect,
		ent.RoleMetadataCreate, ent.RoleMetadataCreateBulk,
		ent.RoleMetadataUpdate, ent.RoleMetadataUpdateOne,
		ent.RoleMetadataDelete,
		predicate.RoleMetadata,
		userV1.RoleMetadata, ent.RoleMetadata,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

// Upsert 插入或更新角色元数据
func (r *RoleMetadataRepo) Upsert(ctx context.Context, data *userV1.RoleMetadata) error {
	err := r.entClient.Client().RoleMetadata.Create().
		SetRoleID(data.GetRoleId()).
		SetNillableIsTemplate(data.IsTemplate).
		SetNillableTemplateFor(data.TemplateFor).
		//SetNillableTemplateVersion(data.TemplateVersion).
		SetNillableLastSyncedVersion(data.LastSyncedVersion).
		SetCustomOverrides(data.CustomOverrides).
		SetCreatedAt(time.Now()).
		OnConflictColumns(
			rolemetadata.FieldRoleID,
		).
		AddTemplateVersion(1).
		SetUpdatedAt(time.Now()).
		Exec(ctx)
	return err
}

// IsExistByRoleID 判断角色元数据是否存在
func (r *RoleMetadataRepo) IsExistByRoleID(ctx context.Context, roleID uint32) (bool, error) {
	count, err := r.entClient.Client().RoleMetadata.Query().
		Where(
			rolemetadata.RoleIDEQ(roleID),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
