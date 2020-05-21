package browserk

// FormHandler will fill the form inputs with context sensitive data
// data can be overridden via config
type FormHandler interface {
	Init() error
	Fill(form *HTMLFormElement)
}
