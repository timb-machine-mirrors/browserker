package plugin

type Loader struct {
}

func NewLoader() *Loader {
	return &Loader{}
}

func LoadPassive() []PassivePlugin {
	return nil
}
