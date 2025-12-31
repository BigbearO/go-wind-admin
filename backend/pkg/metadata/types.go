package metadata

import userV1 "go-wind-admin/api/gen/go/user/service/v1"

const (
	mdOperator = "x-md-operator"
)

// OperatorInfo is a struct for operator metadata.
type OperatorInfo struct {
	UserID          *uint32                `json:"uid,omitempty"`
	TenantID        *uint32                `json:"tid,omitempty"`
	OrgUnitID       *uint32                `json:"ouid,omitempty"`
	IsPlatformAdmin *bool                  `json:"pad,omitempty"`
	DataScope       *userV1.Role_DataScope `json:"ds,omitempty"`
}

func NewOperatorInfo(
	uid uint32,
	tid uint32,
	ouid uint32,
	isPAdmin bool,
	dataScope userV1.Role_DataScope,
) *OperatorInfo {
	return &OperatorInfo{
		UserID:          &uid,
		TenantID:        &tid,
		OrgUnitID:       &ouid,
		IsPlatformAdmin: &isPAdmin,
		DataScope:       &dataScope,
	}
}

func (oi *OperatorInfo) GetUserID() uint32 {
	if oi != nil && oi.UserID != nil {
		return *oi.UserID
	}
	return 0
}

func (oi *OperatorInfo) GetTenantID() uint32 {
	if oi != nil && oi.TenantID != nil {
		return *oi.TenantID
	}
	return 0
}

func (oi *OperatorInfo) GetOrgUnitID() uint32 {
	if oi != nil && oi.OrgUnitID != nil {
		return *oi.OrgUnitID
	}
	return 0
}

func (oi *OperatorInfo) GetIsPlatformAdmin() bool {
	if oi != nil && oi.IsPlatformAdmin != nil {
		return *oi.IsPlatformAdmin
	}
	return false
}

func (oi *OperatorInfo) GetDataScope() userV1.Role_DataScope {
	if oi != nil && oi.DataScope != nil {
		return *oi.DataScope
	}
	return userV1.Role_DATA_SCOPE_UNSPECIFIED
}

func (oi *OperatorInfo) IsSysAdmin() bool {
	return oi.GetIsPlatformAdmin() && oi.GetDataScope() == userV1.Role_ALL
}
