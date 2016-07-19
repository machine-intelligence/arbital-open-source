// Tests for config handling
package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	c, err := load("../../config.yaml")
	if err != nil {
		t.Fatalf("Load() failed: %v\n", err)
	}
	if (reflect.DeepEqual(c, Config{})) {
		t.Fatalf("Load() returned empty Config")
	}
	if (c.MySQL == Config{}.MySQL) {
		t.Fatalf("Load() returned empty Config.MySQL")
	}
	if (c.Site == Config{}.Site) {
		t.Fatalf("Load() returned empty Config.Site")
	}
	if (c.VM == Config{}.VM) {
		t.Fatalf("Load() returned empty Config.Vm")
	}
	if (reflect.DeepEqual(c.Monitoring, Config{}.Monitoring)) {
		t.Fatalf("Load() returned empty Config.Monitoring")
	}
}
