package service

import (
	"context"
	"sort"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	pagination "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	permissionV1 "go-wind-admin/api/gen/go/permission/service/v1"

	"go-wind-admin/pkg/constants"
	"go-wind-admin/pkg/middleware/auth"
	"go-wind-admin/pkg/utils/converter"
)

type PermissionService struct {
	adminV1.PermissionServiceHTTPServer

	log *log.Helper

	permissionRepo *data.PermissionRepo
	menuRepo       *data.MenuRepo
	apiRepo        *data.ApiRepo

	membershipRepo *data.MembershipRepo

	authorizer *data.Authorizer

	menuPermissionConverter *converter.MenuPermissionConverter
	apiPermissionConverter  *converter.ApiPermissionConverter
}

func NewPermissionService(
	ctx *bootstrap.Context,
	permissionRepo *data.PermissionRepo,
	membershipRepo *data.MembershipRepo,
	menuRepo *data.MenuRepo,
	apiRepo *data.ApiRepo,
	authorizer *data.Authorizer,
) *PermissionService {
	svc := &PermissionService{
		log:                     ctx.NewLoggerHelper("permission/service/admin-service"),
		permissionRepo:          permissionRepo,
		membershipRepo:          membershipRepo,
		menuRepo:                menuRepo,
		apiRepo:                 apiRepo,
		authorizer:              authorizer,
		menuPermissionConverter: converter.NewMenuPermissionConverter(),
		apiPermissionConverter:  converter.NewApiPermissionConverter(),
	}

	svc.init()

	return svc
}

func (s *PermissionService) init() {
	ctx := context.Background()
	if count, _ := s.permissionRepo.Count(ctx, []func(s *sql.Selector){}); count == 0 {
		_ = s.createDefaultPermissions(ctx)
	}
	if count, _ := s.apiRepo.Count(ctx, []func(s *sql.Selector){}); count == 0 {
		_, _ = s.SyncApis(ctx, &emptypb.Empty{})
	}
	if count, _ := s.menuRepo.Count(ctx, []func(s *sql.Selector){}); count == 0 {
		_, _ = s.SyncMenus(ctx, &emptypb.Empty{})
	}
}

func (s *PermissionService) List(ctx context.Context, req *pagination.PagingRequest) (*permissionV1.ListPermissionResponse, error) {
	return s.permissionRepo.List(ctx, req)
}

func (s *PermissionService) Get(ctx context.Context, req *permissionV1.GetPermissionRequest) (*permissionV1.Permission, error) {
	return s.permissionRepo.Get(ctx, req)
}

func (s *PermissionService) Create(ctx context.Context, req *permissionV1.CreatePermissionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err = s.permissionRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 重置权限策略
	if err = s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) Update(ctx context.Context, req *permissionV1.UpdatePermissionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.UpdatedBy = trans.Ptr(operator.UserId)

	if err = s.permissionRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 重置权限策略
	if err = s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) Delete(ctx context.Context, req *permissionV1.DeletePermissionRequest) (*emptypb.Empty, error) {
	if err := s.permissionRepo.Delete(ctx, req); err != nil {
		return nil, err
	}

	// 重置权限策略
	if err := s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) SyncApis(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 清理 API 相关权限
	_ = s.permissionRepo.CleanApiPermissions(ctx)

	// 查询所有启用的 API 资源
	apis, err := s.apiRepo.List(ctx, &pagination.PagingRequest{
		NoPaging: trans.Ptr(true),
		Query:    trans.Ptr(`{"status":"ON"}`),
		OrderBy:  []string{"operation"},
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(apis.Items, func(i, j int) bool {
		a, b := apis.Items[i], apis.Items[j]
		if a.GetModule() != b.GetModule() {
			return a.GetModule() < b.GetModule()
		}
		if a.GetPath() != b.GetPath() {
			return a.GetPath() > b.GetPath()
		}
		return a.GetOperation() > b.GetOperation()
	})

	type ApiInfo struct {
		ApiIDs      []uint32
		Description string
	}

	apiInfoMap := make(map[string]ApiInfo)
	for _, api := range apis.Items {
		code := s.apiPermissionConverter.ConvertCodeByOperationID(api.GetOperation())
		if code == "" {
			continue
		}

		if _, exists := apiInfoMap[code]; exists {
			code = s.apiPermissionConverter.ConvertCodeByPath(api.GetMethod(), api.GetPath())
			if code == "" {
				continue
			}
			if _, exists = apiInfoMap[code]; exists {
				s.log.Warnf("SyncApis: duplicate permission code for API %s - %s, skipped", api.GetOperation(), code)
				continue
			}
		}

		if _, exists := apiInfoMap[code]; !exists {
			apiInfoMap[code] = ApiInfo{
				Description: api.GetModuleDescription(),
				ApiIDs:      []uint32{api.GetId()},
			}
		} else {
			info := apiInfoMap[code]
			info.ApiIDs = append(info.ApiIDs, api.GetId())
			apiInfoMap[code] = info
		}

		//s.log.Debugf("SyncApis: prepared permission for API %s - %s", api.GetOperation(), code)
	}

	var permissions []*permissionV1.Permission
	var codes []string
	for code, info := range apiInfoMap {
		permission := &permissionV1.Permission{
			Name:   trans.Ptr(info.Description),
			Code:   trans.Ptr(code),
			Status: trans.Ptr(permissionV1.Permission_ON),
			ApiIds: info.ApiIDs,
		}
		permissions = append(permissions, permission)
		codes = append(codes, code)
	}

	//_ = s.permissionGroupRepo.CleanPermissionsByCodes(ctx, codes)

	if err = s.permissionRepo.BatchCreate(ctx, operator.GetTenantId(), permissions); err != nil {
		s.log.Errorf("batch create api permissions failed: %s", err.Error())
		return nil, err
	}

	// 重置权限策略
	if err = s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) SyncMenus(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	// 获取操作人信息
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 清理菜单相关权限
	_ = s.permissionRepo.CleanMenuPermissions(ctx)

	// 查询所有启用的菜单
	menus, err := s.menuRepo.List(ctx, &pagination.PagingRequest{
		NoPaging: trans.Ptr(true),
		Query:    trans.Ptr(`{"status":"ON"}`),
		OrderBy:  []string{"-id"},
	}, false)
	if err != nil {
		return nil, err
	}

	s.menuPermissionConverter.ComposeMenuPaths(menus.Items)

	sort.SliceStable(menus.Items, func(i, j int) bool {
		return menus.Items[i].GetParentId() < menus.Items[j].GetParentId()
	})

	type MenuInfo struct {
		MenuIDs     []uint32
		Description string
	}

	menuInfoMap := make(map[string]MenuInfo)
	for _, menu := range menus.Items {
		var title string
		//if menu.GetMeta() != nil && menu.GetMeta().GetTitle() != "" {
		//	title = menu.GetMeta().GetTitle()
		//} else {
		//	title = menu.GetName()
		//}
		title = menu.GetName()

		code := s.menuPermissionConverter.ConvertCode(menu.GetPath(), title, menu.GetType())
		if code == "" {
			continue
		}

		if _, exists := menuInfoMap[code]; !exists {
			menuInfoMap[code] = MenuInfo{
				Description: title,
				MenuIDs:     []uint32{menu.GetId()},
			}
		} else {
			info := menuInfoMap[code]
			info.MenuIDs = append(info.MenuIDs, menu.GetId())
			menuInfoMap[code] = info
		}
	}

	//_ = s.permissionGroupRepo.CleanPermissionsByCodes(ctx, codes)

	var permissions []*permissionV1.Permission
	var codes []string
	for code, info := range menuInfoMap {
		permission := &permissionV1.Permission{
			Name:    trans.Ptr(info.Description),
			Code:    trans.Ptr(code),
			Status:  trans.Ptr(permissionV1.Permission_ON),
			MenuIds: info.MenuIDs,
		}
		permissions = append(permissions, permission)
		codes = append(codes, code)
	}

	if err = s.permissionRepo.BatchCreate(ctx, operator.GetTenantId(), permissions); err != nil {
		s.log.Errorf("batch create menu permissions failed: %s", err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) createDefaultPermissions(ctx context.Context) error {
	var err error

	for _, d := range constants.DefaultPermissions {
		if err = s.permissionRepo.Create(ctx, &permissionV1.CreatePermissionRequest{
			Data: d,
		}); err != nil {
			s.log.Errorf("create default permission %s failed: %v", d.GetCode(), err)
			return err
		}
	}

	return nil
}
