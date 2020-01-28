package eclint_test

import (
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"gitlab.com/greut/eclint"
)

func TestOverridingUsingPrefix(t *testing.T) {
	def := &editorconfig.Definition{
		Charset:     "utf-8 bom",
		IndentStyle: "tab",
		IndentSize:  "3",
		TabWidth:    3,
	}

	raw := make(map[string]string)
	raw["@_charset"] = "unset"
	raw["@_indent_style"] = "space"
	raw["@_indent_size"] = "4"
	raw["@_tab_width"] = "4"
	def.Raw = raw

	if err := eclint.OverrideDefinitionUsingPrefix(def, "@_"); err != nil {
		t.Fatal(err)
	}

	if def.Charset != "unset" {
		t.Errorf("charset not changed, got %q", def.Charset)
	}

	if def.IndentStyle != "space" {
		t.Errorf("indent_style not changed, got %q", def.IndentStyle)
	}

	if def.IndentSize != "4" {
		t.Errorf("indent_size not changed, got %q", def.IndentSize)
	}

	if def.TabWidth != 4 {
		t.Errorf("tab_width not changed, got %d", def.TabWidth)
	}
}
