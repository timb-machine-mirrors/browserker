package plugin_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/plugin"
)

const testJSPlugin = "testdata/test_js_plugin.js"

func TestJSPlugin(t *testing.T) {
	s := mock.MakeMockPluginServicer()
	p := plugin.NewJSPluginFromFile(s, testJSPlugin)
	if err := p.Init(); err != nil {
		t.Fatalf("failed to init plugin: %s\n", err)
	}

	if p.ID() != "BR-P-5000" {
		t.Fatalf("plugin ID was invalid got: %v\n", p.ID())
	}
	spew.Dump(p.Name())
}
