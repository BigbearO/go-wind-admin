package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	pagination "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/go-utils/timeutil"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/permission"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

type PermissionRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper          *mapper.CopierMapper[adminV1.Permission, ent.Permission]
	typeConverter   *mapper.EnumTypeConverter[adminV1.Permission_Type, permission.Type]
	statusConverter *mapper.EnumTypeConverter[adminV1.Permission_Status, permission.Status]

	repository *entCrud.Repository[
		ent.PermissionQuery, ent.PermissionSelect,
		ent.PermissionCreate, ent.PermissionCreateBulk,
		ent.PermissionUpdate, ent.PermissionUpdateOne,
		ent.PermissionDelete,
		predicate.Permission,
		adminV1.Permission, ent.Permission,
	]

	permissionApiResourceRepo *PermissionApiResourceRepo
	permissionMenuRepo        *PermissionMenuRepo
}

func NewPermissionRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
	permissionApiResourceRepo *PermissionApiResourceRepo,
	permissionMenuRepo *PermissionMenuRepo,
) *PermissionRepo {
	repo := &PermissionRepo{
		log:                       ctx.NewLoggerHelper("permission/repo/admin-service"),
		entClient:                 entClient,
		mapper:                    mapper.NewCopierMapper[adminV1.Permission, ent.Permission](),
		typeConverter:             mapper.NewEnumTypeConverter[adminV1.Permission_Type, permission.Type](adminV1.Permission_Type_name, adminV1.Permission_Type_value),
		statusConverter:           mapper.NewEnumTypeConverter[adminV1.Permission_Status, permission.Status](adminV1.Permission_Status_name, adminV1.Permission_Status_value),
		permissionApiResourceRepo: permissionApiResourceRepo,
		permissionMenuRepo:        permissionMenuRepo,
	}

	repo.init()

	return repo
}

func (r *PermissionRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PermissionQuery, ent.PermissionSelect,
		ent.PermissionCreate, ent.PermissionCreateBulk,
		ent.PermissionUpdate, ent.PermissionUpdateOne,
		ent.PermissionDelete,
		predicate.Permission,
		adminV1.Permission, ent.Permission,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.typeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *PermissionRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Permission.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, adminV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *PermissionRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListPermissionResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Permission.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &adminV1.ListPermissionResponse{Total: 0, Items: nil}, nil
	}

	return &adminV1.ListPermissionResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *PermissionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Permission.Query().
		Where(permission.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *PermissionRepo) Get(ctx context.Context, req *adminV1.GetPermissionRequest) (*adminV1.Permission, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Permission.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *adminV1.GetPermissionRequest_Id:
		whereCond = append(whereCond, permission.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *PermissionRepo) Create(ctx context.Context, req *adminV1.CreatePermissionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Permission.Create().
		SetNillableCode(req.Data.Code).
		SetName(req.Data.GetName()).
		SetNillablePath(req.Data.Path).
		SetNillableModule(req.Data.Module).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableRemark(req.Data.Remark).
		SetNillableParentID(req.Data.ParentId).
		SetNillableType(r.typeConverter.ToEntity(req.Data.Type)).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetNillableCreatedAt(timeutil.TimestamppbToTime(req.Data.CreatedAt))

	if req.Data.TenantId == nil {
		builder.SetTenantID(req.Data.GetTenantId())
	}
	if req.Data.CreatedAt == nil {
		builder.SetCreatedAt(time.Now())
	}

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert one data failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *PermissionRepo) Update(ctx context.Context, req *adminV1.UpdatePermissionRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &adminV1.CreatePermissionRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Permission.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *adminV1.Permission) {
			builder.
				SetNillableCode(req.Data.Code).
				SetNillableName(req.Data.Name).
				SetNillableModule(req.Data.Module).
				SetNillablePath(req.Data.Path).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableRemark(req.Data.Remark).
				SetNillableParentID(req.Data.ParentId).
				SetNillableType(r.typeConverter.ToEntity(req.Data.Type)).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetNillableUpdatedAt(timeutil.TimestamppbToTime(req.Data.UpdatedAt))

			if req.Data.UpdatedAt == nil {
				builder.SetUpdatedAt(time.Now())
			}
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(permission.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *PermissionRepo) Delete(ctx context.Context, req *adminV1.DeletePermissionRequest) error {
	if req == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Permission.Delete()

	_, err := r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(permission.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete permission failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("delete permission failed")
	}

	return nil
}

// Truncate 清空表数据
func (r *PermissionRepo) Truncate(ctx context.Context) error {
	if _, err := r.entClient.Client().Permission.Delete().Exec(ctx); err != nil {
		r.log.Errorf("failed to truncate permission table: %s", err.Error())
		return adminV1.ErrorInternalServerError("truncate failed")
	}
	return nil
}
