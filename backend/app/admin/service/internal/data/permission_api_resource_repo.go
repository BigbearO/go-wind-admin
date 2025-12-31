package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/permissionapiresource"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

type PermissionApiResourceRepo struct {
	log       *log.Helper
	entClient *entCrud.EntClient[*ent.Client]
}

func NewPermissionApiResourceRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PermissionApiResourceRepo {
	return &PermissionApiResourceRepo{
		log:       ctx.NewLoggerHelper("permission-api-resource/repo/admin-service"),
		entClient: entClient,
	}
}

// CleanApis 清理权限的所有API资源
func (r *PermissionMenuRepo) CleanApis(
	ctx context.Context,
	tx *ent.Tx,
	tenantID uint32,
	permissionIDs []uint32,
) error {
	if _, err := tx.PermissionApiResource.Delete().
		Where(
			permissionapiresource.PermissionIDIn(permissionIDs...),
			permissionapiresource.TenantIDEQ(tenantID),
		).
		Exec(ctx); err != nil {
		err = entCrud.Rollback(tx, err)
		r.log.Errorf("delete old permission apis failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete old permission apis failed")
	}
	return nil
}

// AssignApis 给权限分配API资源
func (r *PermissionApiResourceRepo) AssignApis(ctx context.Context, tx *ent.Tx, tenantID uint32, apis map[uint32]uint32) error {
	if len(apis) == 0 {
		return nil
	}

	for permissionID, apiID := range apis {
		pm := r.entClient.Client().PermissionApiResource.
			Create().
			SetPermissionID(permissionID).
			SetAPIResourceID(apiID).
			SetTenantID(tenantID).
			OnConflict().
			UpdateNewValues()
		if err := pm.Exec(ctx); err != nil {
			err = entCrud.Rollback(tx, err)
			r.log.Errorf("assign permission apis failed: %s", err.Error())
			return adminV1.ErrorInternalServerError("assign permission apis failed")
		}
	}

	return nil
}

// ListApiIDs 列出权限关联的API资源ID列表
func (r *PermissionApiResourceRepo) ListApiIDs(ctx context.Context, tenantID uint32, permissionIDs []uint32) ([]uint32, error) {
	q := r.entClient.Client().PermissionApiResource.
		Query().
		Where(
			permissionapiresource.PermissionIDIn(permissionIDs...),
			permissionapiresource.TenantIDEQ(tenantID),
		)

	intIDs, err := q.
		Select(permissionapiresource.FieldAPIResourceID).
		Ints(ctx)
	if err != nil {
		r.log.Errorf("list permission apis by permission id failed: %s", err.Error())
		return nil, adminV1.ErrorInternalServerError("list permission apis by permission id failed")
	}

	ids := make([]uint32, len(intIDs))
	for i, v := range intIDs {
		ids[i] = uint32(v)
	}
	return ids, nil
}
