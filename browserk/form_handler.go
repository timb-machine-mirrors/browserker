package browserk

type FormHandler interface {
	Init() error
	SetValues(form *HTMLFormElement)
}
