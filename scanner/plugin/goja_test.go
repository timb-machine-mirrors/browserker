package plugin

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
)

func TestGojaLoad(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)

	var pluginScript = `
	var ExecOnce = 1;
	var ExecOncePath = 2;
	var ExecOnceFile = 3;
	var ExecOncePerPage = 4;
	var ExecPerRequest = 5;
	var ExecAlways = 6;

	(function () {
		function Plugin(service) {
			this.service = service;
		}
		Plugin.prototype.Name = function () {
			return "PLUGIN NAME: " + this.service;
		}

		Plugin.prototype.ID = function () {
		    return "BR-P-5000";
		}

		Plugin.prototype.Options = function () {
			return {
				IsolatedRequests: true, // Initiates it's own requests, isolated from a crawl state
				WriteResponses: true, // writes/injects into http/websocket responses
				WriteRequests: true, // writes/injects into http/websocket responses
				WriteJS: true, // writes/injects JS into the browser
				ListenResponses: true, // reads http/websocket responses
				ListenRequests: true, // reads http/websocket requests
				ListenStorage: true, // listens for local/sessionStorage write/read events
				ListenCookies: true, // listens for cookie write events
				ListenConsole: true, // listens for console.log events
				ListenURL: true, // listens for URL change/updates
				ListenJS: true, // listens to JS events
				ExecutionType: ExecAlways // How often/when this plugin executes
			}
		}

		Plugin.prototype.OnEvent = function (evt) {
			console.log(JSON.stringify(evt.PluginEvent));
		}
		
		return Plugin;
	})();
	`

	var createPlugin = `var plugin = new Plugin(service);`

	plugin, err := vm.RunString(pluginScript)
	if err != nil {
		t.Fatalf("error running program: %s\n", err)
	}
	vm.Set("Plugin", plugin)
	vm.Set("service", "blah")

	v, err := vm.RunString(createPlugin)
	if err != nil {
		t.Fatal(err)
	}

	v, err = vm.RunString("plugin.Name()")
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(v)

	v, err = vm.RunString("plugin.ID()")
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(v)

	v, err = vm.RunString("plugin.Options()")
	if err != nil {
		t.Fatal(err)
	}
	opts := &browserk.PluginOpts{}
	vm.ExportTo(v, opts)
	spew.Dump(opts)

	cookies := mock.MakeMockCookies()
	for _, cookie := range cookies {
		c := cookie
		vm.Set("evt", browserk.CookiePluginEvent(nil, "test", nil, c))
		v, err = vm.RunString("plugin.OnEvent(evt)")
		if err != nil {
			t.Fatal(err)
		}
		spew.Dump(v)
	}
}
