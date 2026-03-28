package models

import (
	"time"

	"github.com/tx7do/go-crud/gorm/mixin"
)

// GameAccount 对应表 game_account，游戏账号表
type GameAccount struct {
	mixin.SnowflakeID

	UserAccount *string    `gorm:"column:user_account;type:varchar(100);comment:用户账号"`
	Credential  *string    `gorm:"column:credential;type:varchar(255);comment:用户登录凭证"`
	StartTime   *time.Time `gorm:"column:start_time;type:timestamp;comment:开始时间"`
	ExpireTime  *time.Time `gorm:"column:expire_time;type:timestamp;comment:过期时间"`
	Type        *int32     `gorm:"column:type;type:integer;comment:游戏类型"`
	Enabled     *bool      `gorm:"column:enabled;type:boolean;default:false;comment:是否启用"`
	EmulatorID  *int64     `gorm:"column:emulator_id;type:bigint;comment:分配的模拟器ID"`
	CreateBy    *string    `gorm:"column:create_by;type:varchar(64);comment:创建者"`
	CreateAt    *time.Time `gorm:"column:create_time;type:timestamp;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateBy    *string    `gorm:"column:update_by;type:varchar(64);comment:更新者"`
	UpdateAt    *time.Time `gorm:"column:update_time;type:timestamp;default:CURRENT_TIMESTAMP;comment:更新时间"`
	DelFlag     *bool      `gorm:"column:del_flag;type:boolean;default:false;comment:删除标志"`

	//mixin.TimeAt
	//mixin.OperatorID
	//mixin.Remark
	mixin.TenantID
}

// TableName 指定表名
func (GameAccount) TableName() string {
	return "game_account"
}
