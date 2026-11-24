package testlib

import (
	"fmt"
	"sort"
)

// EmbeddedSchemaIDs are the JSON pointers of every subschema below the root of
// a decoded JSON Schema that carries its own $id: an embedded schema resource,
// which a reference from outside it may not portably point into.
func EmbeddedSchemaIDs(document any) []string {
	var pointers []string
	var walk func(node any, pointer string)
	walk = func(node any, pointer string) {
		switch value := node.(type) {
		case map[string]any:
			if _, hasID := value["$id"]; hasID && pointer != "" {
				pointers = append(pointers, pointer)
			}
			for key, child := range value {
				walk(child, pointer+"/"+key)
			}
		case []any:
			for index, child := range value {
				walk(child, fmt.Sprintf("%s/%d", pointer, index))
			}
		}
	}
	walk(document, "")
	sort.Strings(pointers)
	return pointers
}
