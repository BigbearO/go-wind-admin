package viewer

import (
	userV1 "go-wind-admin/api/gen/go/user/service/v1"
)

// Viewer describes the query/mutation viewer-context.
type Viewer interface {
	// UserId 返回当前用户ID
	UserId() uint32

	// Tenant 返回租户ID
	Tenant() (uint32, bool)

	// --- 身份判定 ---

	// IsSystemAdmin 是否为平台级超级管理员（上帝模式）
	// 对应 JWT 中的 is_padmin = true 且 tenant_id = 0
	IsSystemAdmin() bool

	// IsTenantAdmin 是否为当前租户的管理员
	// 逻辑：!IsSystemAdmin() && DataScope == UNIT_AND_CHILD (且在根节点)
	IsTenantAdmin() bool

	// IsAdmin 泛指具有管理权限（系统管理或租户管理）
	IsAdmin() bool

	// --- 作用域判定 ---

	// IsPlatformContext 当前是否处于平台管理视图（tenant_id == 0）
	IsPlatformContext() bool

	// IsTenantContext 当前是否处于租户业务视图（tenant_id > 0）
	IsTenantContext() bool

	// --- 数据权限（关键新增） ---

	// GetDataScope 返回当前身份的数据权限范围（用于 SQL 拼接）
	GetDataScope() userV1.Role_DataScope // 对应 DataScope 枚举

	// GetOrgUnitId 返回当前身份挂载的组织单元 ID
	GetOrgUnitId() uint32
}

// UserViewer describes a user-viewer.
type UserViewer struct {
	uid       uint32
	tid       uint32
	ouid      uint32
	isPAdmin  bool
	dataScope userV1.Role_DataScope
}

func NewUserViewer(
	uid uint32,
	tid uint32,
	ouid uint32,
	isPAdmin bool,
	dataScope userV1.Role_DataScope,
) UserViewer {
	return UserViewer{
		uid:       uid,
		tid:       tid,
		ouid:      ouid,
		isPAdmin:  isPAdmin,
		dataScope: dataScope,
	}
}

func (v UserViewer) UserId() uint32 {
	return v.uid
}

func (v UserViewer) Tenant() (uint32, bool) {
	return v.tid, v.tid > 0
}

func (v UserViewer) GetDataScope() userV1.Role_DataScope {
	return v.dataScope
}

func (v UserViewer) GetOrgUnitId() uint32 {
	return v.ouid
}

func (v UserViewer) IsAdmin() bool {
	return v.IsSystemAdmin() || v.IsTenantAdmin()
}

func (v UserViewer) IsSystemAdmin() bool {
	return v.isPAdmin && v.tid == 0
}

func (v UserViewer) IsTenantAdmin() bool {
	// 如果是租户视角且拥有本单元及子单元权限（通常租户管理员挂在根节点）
	return v.tid > 0 && v.dataScope == userV1.Role_UNIT_AND_CHILD
}

func (v UserViewer) IsTenantContext() bool {
	return v.tid > 0
}

func (v UserViewer) IsPlatformContext() bool {
	return v.tid == 0
}
