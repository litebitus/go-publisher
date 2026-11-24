package config

import "github.com/invopop/jsonschema"

// inlineSchemaOf reflects value into a schema written in place inside another:
// its nested types inlined, with no $schema, $id or definitions of its own, so
// no reference crosses into a schema resource embedded in the document.
func inlineSchemaOf(value any) *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
	}
	schema := reflector.Reflect(value)
	schema.Version = ""
	schema.ID = ""
	return schema
}

func (a SlackBlock) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

func (a SlackAttachment) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

func (bh Hook) JSONSchema() *jsonschema.Schema {
	type hookAlias Hook
	schema := inlineSchemaOf(&hookAlias{})
	return &jsonschema.Schema{
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
	schema := inlineSchemaOf(&fileAlias{})
	return &jsonschema.Schema{
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
	schema := inlineSchemaOf(&t)
	return &jsonschema.Schema{
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
	schema := inlineSchemaOf(&nixDependencyAlias{})
	return &jsonschema.Schema{
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
	schema := inlineSchemaOf(&pullRequestBaseAlias{})
	return &jsonschema.Schema{
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
	schema := inlineSchemaOf(&homebrewDependencyAlias{})
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			schema,
		},
	}
}
