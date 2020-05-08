package browserk

type AuthService interface {
	Init() error
	Login(c *Context)
}
