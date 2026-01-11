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
	"go-wind-admin/app/admin/service/internal/data/ent/loginauditlog"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

type LoginAuditLogRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[adminV1.LoginAuditLog, ent.LoginAuditLog]

	repository *entCrud.Repository[
		ent.LoginAuditLogQuery, ent.LoginAuditLogSelect,
		ent.LoginAuditLogCreate, ent.LoginAuditLogCreateBulk,
		ent.LoginAuditLogUpdate, ent.LoginAuditLogUpdateOne,
		ent.LoginAuditLogDelete,
		predicate.LoginAuditLog, adminV1.LoginAuditLog, ent.LoginAuditLog,
	]
}

func NewLoginAuditLogRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *LoginAuditLogRepo {
	repo := &LoginAuditLogRepo{
		log:       ctx.NewLoggerHelper("login-audit-log/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[adminV1.LoginAuditLog, ent.LoginAuditLog](),
	}

	repo.init()

	return repo
}

func (r *LoginAuditLogRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.LoginAuditLogQuery, ent.LoginAuditLogSelect,
		ent.LoginAuditLogCreate, ent.LoginAuditLogCreateBulk,
		ent.LoginAuditLogUpdate, ent.LoginAuditLogUpdateOne,
		ent.LoginAuditLogDelete,
		predicate.LoginAuditLog, adminV1.LoginAuditLog, ent.LoginAuditLog,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *LoginAuditLogRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().LoginAuditLog.Query()
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

func (r *LoginAuditLogRepo) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListLoginAuditLogResponse, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().LoginAuditLog.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &adminV1.ListLoginAuditLogResponse{Total: 0, Items: nil}, nil
	}

	return &adminV1.ListLoginAuditLogResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *LoginAuditLogRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().LoginAuditLog.Query().
		Where(loginauditlog.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, adminV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *LoginAuditLogRepo) Get(ctx context.Context, req *adminV1.GetLoginAuditLogRequest) (*adminV1.LoginAuditLog, error) {
	if req == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().LoginAuditLog.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *adminV1.GetLoginAuditLogRequest_Id:
		whereCond = append(whereCond, loginauditlog.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *LoginAuditLogRepo) Create(ctx context.Context, req *adminV1.CreateLoginAuditLogRequest) error {
	if req == nil || req.Data == nil {
		return adminV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().LoginAuditLog.
		Create().
		SetNillableLoginIP(req.Data.LoginIp).
		SetNillableLoginMAC(req.Data.LoginMac).
		SetNillableUserAgent(req.Data.UserAgent).
		SetNillableBrowserName(req.Data.BrowserName).
		SetNillableBrowserVersion(req.Data.BrowserVersion).
		SetNillableClientID(req.Data.ClientId).
		SetNillableClientName(req.Data.ClientName).
		SetNillableOsName(req.Data.OsName).
		SetNillableOsVersion(req.Data.OsVersion).
		SetNillableUserID(req.Data.UserId).
		SetNillableUsername(req.Data.Username).
		SetNillableStatusCode(req.Data.StatusCode).
		SetNillableSuccess(req.Data.Success).
		SetNillableReason(req.Data.Reason).
		SetNillableLocation(req.Data.Location).
		SetNillableLoginTime(timeutil.TimestamppbToTime(req.Data.LoginTime)).
		SetNillableCreatedAt(timeutil.TimestamppbToTime(req.Data.CreatedAt))

	if req.Data.LoginTime == nil {
		builder.SetLoginTime(time.Now())
	}

	if req.Data.CreatedAt == nil {
		builder.SetCreatedAt(time.Now())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert admin login log failed: %s", err.Error())
		return adminV1.ErrorInternalServerError("insert admin login log failed")
	}

	return nil
}
