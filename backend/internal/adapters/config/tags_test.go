package config_test

import (
	"agenticlens/internal/adapters/config"
	"reflect"
	"strings"
	"testing"
)

func TestTagsSharedAndTagsOnlyLocalPrecedenceAndWriter(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, ".agenticlens-config.yaml", fixtureSharedBody(t)+"\ntags:\n  team: core\n  environment: dev\n  quoted: 'true: # text'\n")
	writeConfig(t, root, ".agenticlens-config.local.yaml", "tags:\n  environment: staging\n  local: private-default\n")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tags["team"] != "core" || cfg.Tags["environment"] != "staging" || cfg.Tags["local"] != "private-default" {
		t.Fatalf("local merge: %v", cfg.Tags)
	}
	data, err := config.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private-default") || strings.Contains(string(data), "staging") {
		t.Fatal("effective local tags leaked into shared config writer")
	}
	path := writeConfig(t, t.TempDir(), ".agenticlens-config.yml", string(data))
	roundtrip, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundtrip.Tags, cfg.SharedTags) {
		t.Fatalf("tag config roundtrip: %v / %v", roundtrip.Tags, cfg.SharedTags)
	}
}

func TestTagsInvalidYAMLInBothLayers(t *testing.T) {
	for _, body := range []string{"tags: []\n", "tags: {team: 12}\n", "tags: {team: true}\n", "tags: {12: value}\n", "tags: {team: [one]}\n", "tags: {team: null}\n", "tags: {team: ''}\n", "tags: {' ': value}\n", "tags: {team: one, ' team ': two}\n", "tags: {team: one, team: two}\n"} {
		for _, local := range []bool{false, true} {
			root := t.TempDir()
			shared := fixtureSharedBody(t)
			if local {
				writeConfig(t, root, ".agenticlens-config.local.yaml", body)
			} else {
				shared += "\n" + body
			}
			writeConfig(t, root, ".agenticlens-config.yaml", shared)
			if _, err := config.Load(root); err == nil {
				t.Fatalf("accepted local=%v %q", local, body)
			}
		}
	}
}
