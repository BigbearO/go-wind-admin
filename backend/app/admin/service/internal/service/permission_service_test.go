package service

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"go-wind-admin/app/admin/service/cmd/server/assets"
)

func TestPermissionService_syncWithOpenAPI_EmptyPaths(t *testing.T) {
	// 备份并在结束恢复 OpenApiData
	orig := assets.OpenApiData
	defer func() { assets.OpenApiData = orig }()

	// 构造最小的 PermissionService（只需保证 syncWithOpenAPI 中不会调用 s.log.Fatalf）
	logger := log.NewHelper(log.NewStdLogger(os.Stdout))
	svc := &PermissionService{
		log: logger,
		// 不设置 permissionRepo，因为当 paths 为空时不会进入更新循环
	}

	if err := svc.syncWithOpenAPI(context.Background()); err != nil {
		t.Fatalf("syncWithOpenAPI returned unexpected error: %v", err)
	}
}
