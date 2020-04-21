package plugin

import "gitlab.com/browserker/browserker"

type Loader struct {
}

func NewLoader() *Loader {
	return &Loader{}
}

func LoadPassive() []browserker.PassivePlugin {
	return nil
}
