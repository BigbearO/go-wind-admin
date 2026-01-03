package converter

import (
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/tx7do/go-utils/stringcase"
)

var (
	segVersionRe = regexp.MustCompile(`(?i)^v[0-9]+(?:\.[0-9]+)?$`)
	rpcNameRe    = regexp.MustCompile(`^(?:([A-Za-z0-9]+)Service|([A-Za-z0-9]+))_([A-Za-z]+)([A-Za-z0-9_]*)$`)
)

type ApiPermissionConverter struct {
}

func NewApiPermissionConverter() *ApiPermissionConverter {
	return &ApiPermissionConverter{}
}

// ConvertCodeByOperationID 通过 operationID 生成 resource:action 风格的 code（如 users:delete, users:list）
func (c *ApiPermissionConverter) ConvertCodeByOperationID(operationID string) string {
	service, action, name, ok := c.splitOperationID(operationID)
	if !ok {
		return ""
	}

	resource := stringcase.KebabCase(service)
	if name != "" {
		resource = resource + ":" + stringcase.KebabCase(name)
	}

	return inflection.Singular(resource) + ":" + stringcase.KebabCase(action)
}

// ConvertCodeByPath 通过 HTTP 方法和路径生成 resource:action 风格的 code（如 users:delete, users:list）
func (c *ApiPermissionConverter) ConvertCodeByPath(method, path string) string {
	resource := c.pathToResource(path)
	action := c.methodToAction(method, path)
	return inflection.Singular(resource) + ":" + action
}

// methodToAction 将 HTTP 方法转换为动作字符串
func (c *ApiPermissionConverter) methodToAction(method string, path string) string {
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
	return action
}

// pathToResource 从路径中解析资源标识符
func (c *ApiPermissionConverter) pathToResource(path string) string {
	if path == "" {
		return ""
	}

	path = c.stripVersionPrefix(path)
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

// splitOperationID 分割 operationID 为 resource 和 action
func (c *ApiPermissionConverter) splitOperationID(operationID string) (service, action, name string, ok bool) {
	if operationID == "" {
		return "", "", "", false
	}
	m := rpcNameRe.FindStringSubmatch(operationID)
	if len(m) < 5 {
		return "", "", "", false
	}
	// 取非空的 service 捕获组
	if m[1] != "" {
		service = m[1]
	} else if m[2] != "" {
		service = m[2]
	} else {
		return "", "", "", false
	}

	actionRaw := m[3]
	nameRaw := m[4]

	if nameRaw == "" {
		// 尝试从 actionRaw 中按驼峰边界拆分出 name
		a, n := c.splitCamelAfterFirstWord(actionRaw)
		action = a
		name = n
	} else {
		action = actionRaw
		name = nameRaw
	}
	return service, action, name, true
}

// splitCamelAfterFirstWord 将驼峰字符串拆为第一个大写词和剩余部分。
// 例如 ListTaskTypeName -> ("List","TaskTypeName")，List -> ("List","")
func (c *ApiPermissionConverter) splitCamelAfterFirstWord(s string) (first, rest string) {
	if s == "" {
		return "", ""
	}
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if ch >= 'A' && ch <= 'Z' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// stripVersionPrefix 移除路径开头的 /api/ 和 /vN/
// 例如:
//
//	/api/v1/admin/settings -> admin/settings
//	/admin/api/v1/settings -> settings
//	/v1/users             -> users
//	/v2                  -> ""
func (c *ApiPermissionConverter) stripVersionPrefix(p string) string {
	if strings.TrimSpace(p) == "" {
		return ""
	}
	p = strings.TrimSpace(p)
	p = strings.Trim(p, "/")
	if p == "" {
		return ""
	}

	parts := strings.Split(p, "/")

	// 去掉开头的 "api"
	if len(parts) > 0 && strings.EqualFold(parts[0], "api") {
		parts = parts[1:]
	}

	// 找到第一个版本段
	idx := -1
	for i, seg := range parts {
		if segVersionRe.MatchString(seg) {
			idx = i
			break
		}
	}

	// 如果找到了且位置在前两段（索引 0 或 1），移除从头到该版本段（含）
	if idx != -1 && idx <= 1 {
		if idx+1 <= len(parts) {
			parts = parts[idx+1:]
		} else {
			parts = nil
		}
	}

	// 过滤空段及残余的 "api"
	var out []string
	for _, seg := range parts {
		seg = strings.TrimSpace(seg)
		if seg == "" || strings.EqualFold(seg, "api") {
			continue
		}
		out = append(out, seg)
	}

	return strings.Join(out, "/")
}
