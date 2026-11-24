package config

//go:generate go run ../../internal/fielddescriptions/generate github.com/litebitus/go-publisher/pkg/config field_descriptions.go

import (
	"reflect"

	"github.com/invopop/jsonschema"
)

// inlineSchemaOf reflects value, a pointer to an alias of the type named
// describedAs, into a schema written in place inside another: its nested types
// inlined, with no $schema, $id or definitions of its own, so no reference
// crosses into a schema resource embedded in the document. Its descriptions are
// describedAs's.
func inlineSchemaOf(value any, describedAs string) *jsonschema.Schema {
	aliasType := reflect.TypeOf(value).Elem()
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
		CommentMap:     FieldDescriptions,
		LookupComment: func(described reflect.Type, field string) string {
			if described != aliasType {
				return ""
			}
			name := aliasType.PkgPath() + "." + describedAs
			if field != "" {
				name += "." + field
			}
			return FieldDescriptions[name]
		},
	}
	schema := reflector.Reflect(value)
	schema.Version = ""
	schema.ID = ""
	return schema
}

// typeDescription is the doc comment of value's type, for a type whose schema
// its own JSONSchema method writes.
func typeDescription(value any) string {
	described := reflect.TypeOf(value)
	return FieldDescriptions[described.PkgPath()+"."+described.Name()]
}

func (a SlackBlock) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Description:          typeDescription(a),
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

func (a SlackAttachment) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Description:          typeDescription(a),
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

func (bh Hook) JSONSchema() *jsonschema.Schema {
	type hookAlias Hook
	schema := inlineSchemaOf(&hookAlias{}, "Hook")
	return &jsonschema.Schema{
		Description: typeDescription(bh),
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}

func (f File) JSONSchema() *jsonschema.Schema {
	type fileAlias File
	schema := inlineSchemaOf(&fileAlias{}, "File")
	return &jsonschema.Schema{
		Description: typeDescription(f),
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}

func (bhc Hooks) JSONSchema() *jsonschema.Schema {
	var t Hook
	schema := inlineSchemaOf(&t, "Hook")
	return &jsonschema.Schema{
		Description: typeDescription(bhc),
		OneOf: []*jsonschema.Schema{{
			Type: "string",
		}, {
			Type:  "array",
			Items: schema,
		}},
	}
}

func (a FlagArray) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Description: typeDescription(a),
		OneOf: []*jsonschema.Schema{{
			Type: "string",
		}, {
			Type: "array",
			Items: &jsonschema.Schema{
				Type: "string",
			},
		}},
	}
}

func (a StringArray) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Description: typeDescription(a),
		OneOf: []*jsonschema.Schema{{
			Type: "string",
		}, {
			Type: "array",
			Items: &jsonschema.Schema{
				Type: "string",
			},
		}},
	}
}

func (a NixDependency) JSONSchema() *jsonschema.Schema {
	type nixDependencyAlias NixDependency
	schema := inlineSchemaOf(&nixDependencyAlias{}, "NixDependency")
	return &jsonschema.Schema{
		Description: typeDescription(a),
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}

func (a PullRequestBase) JSONSchema() *jsonschema.Schema {
	type pullRequestBaseAlias PullRequestBase
	schema := inlineSchemaOf(&pullRequestBaseAlias{}, "PullRequestBase")
	return &jsonschema.Schema{
		Description: typeDescription(a),
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}

func (a HomebrewDependency) JSONSchema() *jsonschema.Schema {
	type homebrewDependencyAlias HomebrewDependency
	schema := inlineSchemaOf(&homebrewDependencyAlias{}, "HomebrewDependency")
	return &jsonschema.Schema{
		Description: typeDescription(a),
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}
