package service

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	permissionV1 "go-wind-admin/api/gen/go/permission/service/v1"
)

type OperationAuditLogService struct {
	adminV1.OperationAuditLogServiceHTTPServer

	log *log.Helper

	operationLogRepo *data.OperationAuditLogRepo
	apiRepo          *data.ApiRepo

	apis     []*permissionV1.Api
	apiMutex sync.RWMutex
}

func NewOperationAuditLogService(
	ctx *bootstrap.Context,
	operationLogRepo *data.OperationAuditLogRepo,
	apiRepo *data.ApiRepo,
) *OperationAuditLogService {
	return &OperationAuditLogService{
		log:              ctx.NewLoggerHelper("operation-audit-log/service/admin-service"),
		operationLogRepo: operationLogRepo,
		apiRepo:          apiRepo,
	}
}

func (s *OperationAuditLogService) queryApis(ctx context.Context, path, method string) (*permissionV1.Api, error) {
	if len(s.apis) == 0 {
		s.apiMutex.Lock()
		apis, err := s.apiRepo.List(ctx, &pagination.PagingRequest{
			NoPaging: trans.Ptr(true),
		})
		if err != nil {
			s.apiMutex.Unlock()
			return nil, err
		}
		s.apis = apis.Items
		s.apiMutex.Unlock()
	}

	if len(s.apis) == 0 {
		return nil, adminV1.ErrorNotFound("no apis found")
	}

	for _, api := range s.apis {
		if api.GetPath() == path && api.GetMethod() == method {
			return api, nil
		}
	}

	return nil, nil
}

func (s *OperationAuditLogService) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListOperationAuditLogResponse, error) {
	resp, err := s.operationLogRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(resp.Items); i++ {
		l := resp.Items[i]
		if l == nil {
			continue
		}
		a, _ := s.queryApis(ctx, l.GetPath(), l.GetMethod())
		if a != nil {
			l.ApiDescription = a.Description
			l.ApiModule = a.ModuleDescription
		}
	}

	return resp, nil
}

func (s *OperationAuditLogService) Get(ctx context.Context, req *adminV1.GetOperationAuditLogRequest) (*adminV1.OperationAuditLog, error) {
	resp, err := s.operationLogRepo.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	a, _ := s.queryApis(ctx, resp.GetPath(), resp.GetMethod())
	if a != nil {
		resp.ApiDescription = a.Description
		resp.ApiModule = a.ModuleDescription
	}

	return resp, nil
}

func (s *OperationAuditLogService) Create(ctx context.Context, req *adminV1.CreateOperationAuditLogRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.operationLogRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
