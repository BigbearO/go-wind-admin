package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// RoleMetadata 角色元数据（模板标记/覆盖项/版本控制）
type RoleMetadata struct {
	ent.Schema
}

func (RoleMetadata) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_role_metadata",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("角色元数据"),
	}
}

// Fields of the RoleMetadata.
func (RoleMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("role_id").
			Comment("角色ID").
			Optional().
			Nillable(),

		field.Bool("is_template").
			Comment("是否是模版").
			Optional().
			Default(false).
			Nillable(),

		field.String("template_for").
			Comment("模板适用对象").
			Optional().
			Nillable(),

		field.Int32("template_version").
			Comment("模板版本号").
			Default(1).
			Optional().
			Nillable(),

		field.Int32("last_synced_version").
			Comment("上次同步的版本号").
			Default(0).
			Optional().
			Nillable(),

		field.Strings("custom_overrides").
			Comment("自定义覆盖项"),
	}
}

// Mixin of the RoleMetadata.
func (RoleMetadata) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
	}
}

// Indexes of the RoleMetadata.
func (RoleMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id").Unique(),
		index.Fields("is_template", "template_for"),
		index.Fields("template_for", "template_version"),
	}
}
