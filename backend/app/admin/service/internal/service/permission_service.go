package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"entgo.io/ent/dialect/sql"
	"github.com/getkin/kin-openapi/openapi3"
	pagination "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/cmd/server/assets"
	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

type PermissionService struct {
	adminV1.PermissionServiceHTTPServer

	log *log.Helper

	permissionRepo *data.PermissionRepo
	membershipRepo *data.MembershipRepo

	authorizer  *data.Authorizer
	routeWalker RouteWalker
}

func NewPermissionService(
	ctx *bootstrap.Context,
	permissionRepo *data.PermissionRepo,
	membershipRepo *data.MembershipRepo,
	authorizer *data.Authorizer,
) *PermissionService {
	svc := &PermissionService{
		log:            ctx.NewLoggerHelper("permission/service/admin-service"),
		permissionRepo: permissionRepo,
		membershipRepo: membershipRepo,
		authorizer:     authorizer,
	}

	svc.init()

	return svc
}

func (s *PermissionService) init() {
	ctx := context.Background()
	if count, _ := s.permissionRepo.Count(ctx, []func(s *sql.Selector){}); count == 0 {
		_, _ = s.SyncPermissions(ctx, &emptypb.Empty{})
	}
}

func (s *PermissionService) RegisterRouteWalker(routeWalker RouteWalker) {
	s.routeWalker = routeWalker
}

func (s *PermissionService) List(ctx context.Context, req *pagination.PagingRequest) (*adminV1.ListPermissionResponse, error) {
	return s.permissionRepo.List(ctx, req)
}

func (s *PermissionService) Get(ctx context.Context, req *adminV1.GetPermissionRequest) (*adminV1.Permission, error) {
	return s.permissionRepo.Get(ctx, req)
}

func (s *PermissionService) Create(ctx context.Context, req *adminV1.CreatePermissionRequest) (*emptypb.Empty, error) {
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

func (s *PermissionService) Update(ctx context.Context, req *adminV1.UpdatePermissionRequest) (*emptypb.Empty, error) {
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

func (s *PermissionService) Delete(ctx context.Context, req *adminV1.DeletePermissionRequest) (*emptypb.Empty, error) {
	if err := s.permissionRepo.Delete(ctx, req); err != nil {
		return nil, err
	}

	// 重置权限策略
	if err := s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PermissionService) SyncPermissions(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	_ = s.permissionRepo.Truncate(ctx)

	if err := s.syncWithOpenAPI(ctx); err != nil {
		return nil, err
	}

	// 重置权限策略
	if err := s.authorizer.ResetPolicies(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// syncWithOpenAPI 使用 OpenAPI 文档同步 API 资源
func (s *PermissionService) syncWithOpenAPI(ctx context.Context) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(assets.OpenApiData)
	if err != nil {
		s.log.Fatalf("加载 OpenAPI 文档失败: %v", err)
		return adminV1.ErrorInternalServerError("load OpenAPI document failed")
	}

	if doc == nil {
		s.log.Fatal("OpenAPI 文档为空")
		return adminV1.ErrorInternalServerError("OpenAPI document is nil")
	}
	if doc.Paths == nil {
		s.log.Fatal("OpenAPI 文档的路径为空")
		return adminV1.ErrorInternalServerError("OpenAPI document paths is nil")
	}

	var count uint32 = 0
	var apiResourceList []*adminV1.Permission

	// 遍历所有路径和操作
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {

			var module string
			//var moduleDescription string
			if len(operation.Tags) > 0 {
				tag := doc.Tags.Get(operation.Tags[0])
				if tag != nil {
					module = tag.Name
					//moduleDescription = tag.Description
				}
			}

			count++

			resource := ParseResourceFromPath(path)
			code := BuildCode(method, path, operation.OperationID, module, resource)
			//fmt.Println(moduleDescription, resource, code)

			apiResourceList = append(apiResourceList, &adminV1.Permission{
				Id:   trans.Ptr(count),
				Name: trans.Ptr(operation.Description),
				Code: trans.Ptr(code),
				//Path:     trans.Ptr(path),
				//Resource: trans.Ptr(resource),
				//Method:   trans.Ptr(method),
				Type:   trans.Ptr(adminV1.Permission_API),
				Status: trans.Ptr(adminV1.Permission_ON),
			})
		}
	}

	for i, res := range apiResourceList {
		res.Id = trans.Ptr(uint32(i + 1))
		_ = s.permissionRepo.Update(ctx, &adminV1.UpdatePermissionRequest{
			AllowMissing: trans.Ptr(true),
			Data:         res,
		})
	}

	return nil
}

var (
	rePrefix   = regexp.MustCompile(`(?i)^(?:[^/]+/)?v\d+(?:\.\d+)?(?:/|$)`)
	reNonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)
)

// BuildCode 生成 resource:action 风格的 code（如 users:delete, users:list）
func BuildCode(method, path, operationID, tag, resource string) string {
	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "{}")
		s = strings.Trim(s, "_:")
		s = strings.ReplaceAll(s, "_:", ":")
		s = strings.ReplaceAll(s, ":_", ":")
		return strings.ToLower(s)
	}

	h := md5.Sum([]byte(method + "|" + path + "|" + operationID + "|" + tag + "|" + resource))
	res := ""
	if resource != "" {
		res = normalize(resource)
	}
	if res == "" {
		parts := strings.FieldsFunc(operationID, func(r rune) bool {
			return r == '_' || r == '-' || r == '.'
		})
		if len(parts) > 0 {
			res = normalize(parts[len(parts)-1])
		}
		if res == "" {
			res = hex.EncodeToString(h[:])[:8]
		}
	}

	action := "get"
	switch strings.ToUpper(method) {
	case "GET":
		if strings.Contains(path, "{") {
			action = "get"
		} else {
			action = "list"
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = strings.ToLower(method)
	}

	return res + ":" + action
}

// ParseResourceFromPath 从路径中解析资源标识符
func ParseResourceFromPath(path string) string {
	if path == "" {
		return ""
	}
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	path = rePrefix.ReplaceAllString(path, "")
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	var segs []string

	for _, p := range parts {
		raw := strings.TrimSpace(p)
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
			continue
		}
		clean := strings.Trim(raw, "{}")
		clean = reNonAlnum.ReplaceAllString(clean, "_")
		clean = strings.Trim(clean, "_")
		clean = strings.ToLower(clean)
		if clean == "" {
			continue
		}
		if clean == "api" || (strings.HasPrefix(clean, "v") && len(clean) <= 3) {
			continue
		}
		segs = append(segs, clean)
	}

	if len(segs) == 0 {
		return ""
	}

	return strings.Join(segs, ":")
}
