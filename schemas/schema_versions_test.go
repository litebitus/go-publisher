package schemas

import (
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"
)

// Every minor's file is identified by its own name and titled with its version.
func TestEveryMinorSchemaIsIdentifiedByItsVersion(t *testing.T) {
	minorSchemas, err := readMinorSchemas(schemaFiles)
	if err != nil || len(minorSchemas) == 0 {
		t.Fatalf("read %d schemas (%v)", len(minorSchemas), err)
	}
	for _, schema := range minorSchemas {
		text, err := schemaFiles.ReadFile(schema.fileName)
		if err != nil {
			t.Fatal(err)
		}
		var read struct {
			ID    string `json:"$id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(text, &read); err != nil {
			t.Fatalf("%s: %v", schema.fileName, err)
		}
		if read.ID != "https://github.com/litebitus/go-publisher/schemas/"+schema.fileName ||
			read.Title != "go-publisher configuration v"+schema.version() {
			t.Errorf("%s: $id %s, title %s", schema.fileName, read.ID, read.Title)
		}
	}
}

// A release attaches every minor, and each major as its newest minor.
func TestTheReleaseAssetsAreEveryMinorAndEachMajorAsItsNewestMinor(t *testing.T) {
	assets, err := ReleaseAssets()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go-publisher-v2.0.schema.json", "go-publisher-v2.1.schema.json"} {
		if assets[name] == nil {
			t.Errorf("no asset %s", name)
		}
	}
	minorSchemas, _ := readMinorSchemas(schemaFiles)
	newest := minorSchemas[len(minorSchemas)-1]
	floating := assets[fmt.Sprintf("go-publisher-v%d.schema.json", newest.major)]
	if floating == nil || string(floating) != string(assets[newest.fileName]) {
		t.Errorf("the major's asset is not its newest minor %s", newest.fileName)
	}
}

// A stated major reads as its newest minor; a major with no schema is refused.
func TestAMajorReadsAsItsNewestMinor(t *testing.T) {
	text, minorVersion, err := SchemaFor(2)
	if err != nil || minorVersion != "2.1" || string(text) == "" {
		t.Errorf("version 2 reads as %s (%v)", minorVersion, err)
	}
	if _, _, err := SchemaFor(1); err == nil || err.Error() != "go-publisher has no schema for version 1" {
		t.Errorf("a major with no schema: err = %v", err)
	}
}

// Files are ordered by version, a minor of ten after a minor of nine, and a
// file not named for a version is refused.
func TestSchemaFilesAreReadInVersionOrder(t *testing.T) {
	read, err := readMinorSchemas(fstest.MapFS{
		"go-publisher-v2.10.schema.json": {},
		"go-publisher-v2.9.schema.json":  {},
		"go-publisher-v3.0.schema.json":  {},
	})
	if err != nil {
		t.Fatal(err)
	}
	var versions []string
	for _, schema := range read {
		versions = append(versions, schema.version())
	}
	if fmt.Sprint(versions) != "[2.9 2.10 3.0]" {
		t.Errorf("versions = %v", versions)
	}
	if _, err := readMinorSchemas(fstest.MapFS{"mock.schema.json": {}}); err == nil {
		t.Error("a misnamed schema file was read")
	}
}
