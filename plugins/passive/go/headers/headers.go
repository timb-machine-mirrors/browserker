package headers

import (
	"fmt"

	"gitlab.com/simpscan/scanner/plugin"
)

type Headers struct {
	Config *plugin.PassiveConfig
}

func (h *Headers) OnRequest(data []byte) {
	fmt.Printf("This is a plugin test")
}

var HeaderCheck Headers
