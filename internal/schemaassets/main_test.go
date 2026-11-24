package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTheAssetsAreEveryMinorAndEachMajorAsItsNewestMinor(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "mock-assets")
	if err := writeAssets([]string{folder}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(filepath.Join(folder, "go-publisher-v2.0.schema.json")); err != nil {
		t.Fatal(err)
	}
	minor, err := os.ReadFile(filepath.Join(folder, "go-publisher-v2.1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	major, err := os.ReadFile(filepath.Join(folder, "go-publisher-v2.schema.json"))
	if err != nil || string(major) != string(minor) {
		t.Errorf("the major's asset is not its newest minor (%v)", err)
	}
}

func TestTheAssetsNeedOneFolderTheyCanBeWrittenTo(t *testing.T) {
	if err := writeAssets(nil); err == nil || err.Error() != "usage: schemaassets <folder>" {
		t.Errorf("err = %v", err)
	}
	blocked := filepath.Join(t.TempDir(), "mock-file")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeAssets([]string{blocked}); err == nil {
		t.Error("the assets were written under a file")
	}
}
