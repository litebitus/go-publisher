// Package schemas holds the configuration's JSON Schema, one file per minor version.
package schemas

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
)

//go:embed go-publisher-v*.schema.json
var schemaFiles embed.FS

var minorFileName = regexp.MustCompile(`^go-publisher-v(\d+)\.(\d+)\.schema\.json$`)

// minorSchema is one minor version's schema file.
type minorSchema struct {
	major, minor int
	fileName     string
}

func (schema minorSchema) version() string { return fmt.Sprintf("%d.%d", schema.major, schema.minor) }

func readMinorSchemas(files fs.FS) ([]minorSchema, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to list the embedded schemas: %w", err)
	}
	var read []minorSchema
	for _, entry := range entries {
		named := minorFileName.FindStringSubmatch(entry.Name())
		if named == nil {
			return nil, fmt.Errorf("%s is not named go-publisher-v<major>.<minor>.schema.json", entry.Name())
		}
		major, _ := strconv.Atoi(named[1])
		minor, _ := strconv.Atoi(named[2])
		read = append(read, minorSchema{major: major, minor: minor, fileName: entry.Name()})
	}
	slices.SortFunc(read, func(left, right minorSchema) int {
		if left.major != right.major {
			return left.major - right.major
		}
		return left.minor - right.minor
	})
	return read, nil
}

// ReleaseAssets are the schema files a release attaches, by asset name: every minor's file, and each major's as its newest minor, for a modeline to name.
func ReleaseAssets() (map[string][]byte, error) {
	minorSchemas, err := readMinorSchemas(schemaFiles)
	if err != nil {
		return nil, err
	}
	assets := map[string][]byte{}
	for _, schema := range minorSchemas {
		text, err := schemaFiles.ReadFile(schema.fileName)
		if err != nil {
			return nil, err
		}
		assets[schema.fileName] = text
		assets[fmt.Sprintf("go-publisher-v%d.schema.json", schema.major)] = text
	}
	return assets, nil
}
