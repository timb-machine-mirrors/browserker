package plugin

import "gitlab.com/browserker/browserker"

type Manager struct {
}

func NewLoader() *Manager {
	return &Manager{}
}

func LoadPassive() []browserker.Plugin {
	return nil
}
