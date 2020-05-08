package plugin

type Manager struct {
	dispatch *EventDispatcher
}

func New() *Manager {
	return &Manager{dispatch: NewEventDispatcher()}
}
