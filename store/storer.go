package store

type Storer interface {
	Init() error
	Load(path string) error
	Close() error
}
