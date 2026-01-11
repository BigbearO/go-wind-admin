package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// LoginPolicy holds the schema definition for the LoginPolicy entity.
type LoginPolicy struct {
	ent.Schema
}

func (LoginPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_login_policies",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("登录策略表"),
	}
}

// Fields of the LoginPolicy.
func (LoginPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("target_id").
			Comment("目标用户ID").
			Optional().
			Nillable(),

		field.String("value").
			Comment("限制值（如IP地址、MAC地址或地区代码）").
			Optional().
			Nillable(),

		field.String("reason").
			Comment("限制原因").
			Optional().
			Nillable(),

		field.Enum("type").
			Comment("限制类型").
			NamedValues(
				"Blacklist", "BLACK_LIST",
				"Whitelist", "WHITE_LIST",
			).
			Default("BLACK_LIST").
			Optional().
			Nillable(),

		field.Enum("method").
			Comment("限制方式").
			NamedValues(
				"Ip", "IP",
				"Mac", "MAC",
				"Region", "REGION",
				"Time", "TIME",
				"Device", "DEVICE",
			).
			Default("IP").
			Optional().
			Nillable(),
	}
}

// Indexes of the LoginPolicy.
func (LoginPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("target_id", "type", "method").Unique().
			StorageKey("idx_sys_login_policy_target_type_method"),
	}
}

// Mixin of the LoginPolicy.
func (LoginPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID{},
	}
}
