package inject

type Injector interface {
	Get() []byte
}
