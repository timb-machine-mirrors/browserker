package plugin

import "gitlab.com/browserker/browserk"

type Manager struct {
}

func NewLoader() *Manager {
	return &Manager{}
}

func LoadPassive() []browserk.Plugin {
	return nil
}
