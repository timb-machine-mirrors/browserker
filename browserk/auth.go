package browserk

// AuthService handles logging in and checking login state throughout a scan/crawl
type AuthService interface {
	Init() error
	Login(c *Context)
	MustLogin() bool
}
