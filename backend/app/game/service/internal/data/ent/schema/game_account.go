package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type GameAccount struct {
	ent.Schema
}

func (GameAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "game_account",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("游戏账号表"),
	}
}

func (GameAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_account").
			Comment("用户账号").
			MaxLen(100).
			Optional().
			Nillable(),

		field.String("credential").
			Comment("用户登录凭证").
			MaxLen(255).
			Optional().
			Nillable(),

		field.Time("start_time").
			Comment("开始时间").
			Optional().
			Nillable(),

		field.Time("expire_time").
			Comment("过期时间").
			Optional().
			Nillable(),

		field.Int32("type").
			Comment("游戏类型").
			Optional().
			Nillable(),

		field.Bool("enabled").
			Comment("是否启用").
			Default(false),

		field.Int64("emulator_id").
			Comment("分配的模拟器ID").
			Optional().
			Nillable(),
	}
}

func (GameAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.OperatorID{},
		mixin.TimeAt{},
		mixin.TenantID[uint32]{},
	}
}

func (GameAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_account").Unique().StorageKey("idx_game_account_tenant_user_account"),
		index.Fields("tenant_id", "type").StorageKey("idx_game_account_tenant_type"),
		index.Fields("tenant_id", "enabled").StorageKey("idx_game_account_tenant_enabled"),
	}
}
