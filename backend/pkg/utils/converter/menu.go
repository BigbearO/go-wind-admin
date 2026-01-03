package converter

import (
	"strings"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"

	"github.com/jinzhu/inflection"
	"github.com/tx7do/go-utils/trans"
)

type MenuPermissionConverter struct {
}

func NewMenuPermissionConverter() *MenuPermissionConverter {
	return &MenuPermissionConverter{}
}

// MenuTypeToPermissionType 将 Menu_Type 转换为 Permission_Type
func (c *MenuPermissionConverter) MenuTypeToPermissionType(typ adminV1.Menu_Type) adminV1.Permission_Type {
	switch typ {
	case adminV1.Menu_CATALOG:
		return adminV1.Permission_CATALOG
	case adminV1.Menu_MENU:
		return adminV1.Permission_MENU
	case adminV1.Menu_BUTTON, adminV1.Menu_EMBEDDED, adminV1.Menu_LINK:
		return adminV1.Permission_BUTTON
	default:
		return adminV1.Permission_OTHER
	}
}

// ConvertCode 将菜单的完整路径和类型转换为权限代码
func (c *MenuPermissionConverter) ConvertCode(fullPath string, typ adminV1.Menu_Type) string {
	path := strings.TrimSpace(fullPath)
	if path == "" {
		return ""
	}

	// 移除掉路径前后的斜杠
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}

	// 将路径段用 ':' 连接，作为权限主体
	permBase := strings.ReplaceAll(path, "/", ":")
	permBase = inflection.Singular(permBase)

	// 根据菜单类型，决定是否添加动作后缀
	action := c.typeToAction(typ)
	if action == "" {
		return permBase
	}

	return permBase + ":" + action
}

// ComposeMenuPaths 递归拼接 menus 中每个菜单的 path 并写回菜单的 Path 字段（*string）。
// - menus: 待处理的菜单切片（会就地修改）
// 行为说明：
// 1. 使用 id->menu 映射快速查找父节点。
// 2. 用递归 + memoization 计算每个节点的完整 path（去除两端斜杠并用 '/' 连接）。
// 3. 若父节点 id 为 0 或父节点不存在，则视为根路径（仅使用自身 path 部分）。
// 4. 若出现自引用或循环，函数会将该节点视为只使用自身 path。
func (c *MenuPermissionConverter) ComposeMenuPaths(menus []*adminV1.Menu) {
	// 建立 id -> menu 映射
	m := make(map[uint32]*adminV1.Menu, len(menus))
	for _, mi := range menus {
		m[mi.GetId()] = mi
	}

	// 记忆计算结果：id -> fullPath
	memo := make(map[uint32]string, len(menus))

	var compute func(id uint32, seen map[uint32]bool) string
	compute = func(id uint32, seen map[uint32]bool) string {
		// 已计算
		if v, ok := memo[id]; ok {
			return v
		}
		// 循环检测
		if seen[id] {
			memo[id] = strings.Trim(m[id].GetPath(), "/")
			return memo[id]
		}
		menu, ok := m[id]
		if !ok {
			memo[id] = ""
			return ""
		}

		seen[id] = true
		defer delete(seen, id)

		//part := strings.Trim(menu.GetPath(), "/")
		part := menu.GetPath()
		parentId := menu.GetParentId()
		// 根节点或无父节点
		if parentId == 0 || parentId == id {
			memo[id] = part
			return memo[id]
		}
		parent, ok := m[parentId]
		if !ok {
			memo[id] = part
			return memo[id]
		}

		parentFull := compute(parent.GetId(), seen)
		var fullPath string
		switch {
		case parentFull == "":
			fullPath = part
		case part == "":
			fullPath = parentFull
		default:
			fullPath = parentFull + "/" + part
		}
		memo[id] = fullPath
		return fullPath
	}

	// 为每个菜单计算并写回 Path 字段
	for _, menu := range menus {
		id := menu.GetId()
		fullPath := compute(id, map[uint32]bool{})
		//log.Infof("Menu ID %d full path: %s", id, fullPath)
		// 写回为指针字符串
		menu.Path = trans.Ptr(fullPath)
	}
}

// typeToAction 将 Menu_Type 转换为 action 字符串
func (c *MenuPermissionConverter) typeToAction(typ adminV1.Menu_Type) string {
	switch typ {
	case adminV1.Menu_CATALOG:
		return ""
	case adminV1.Menu_MENU:
		return "view"
	case adminV1.Menu_BUTTON:
		return ""
	case adminV1.Menu_EMBEDDED:
		return "embed"
	case adminV1.Menu_LINK:
		return "link"
	default:
		return ""
	}
}
