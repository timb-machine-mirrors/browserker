package browser

import (
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/wirepair/gcd"
)

type LocalLeaser struct {
	browserLock    sync.RWMutex
	browsers       map[string]*gcd.Gcd
	browserTimeout time.Duration
	tmp            string
}

func NewLocalLeaser() *LocalLeaser {
	s := &LocalLeaser{
		browserLock:    sync.RWMutex{},
		browserTimeout: time.Second * 30,
		browsers:       make(map[string]*gcd.Gcd),
	}
	return s
}

func (s *LocalLeaser) Acquire() (string, error) {
	b := gcd.NewChromeDebugger()
	b.DeleteProfileOnExit()

	chrome, tmp := FindChrome()
	profileDir := randProfile(tmp)
	s.tmp = tmp
	port := randPort()

	b.AddFlags(startupFlags)
	if err := b.StartProcess(chrome, profileDir, port); err != nil {
		return "", err
	}
	s.browserLock.Lock()
	s.browsers[port] = b
	s.browserLock.Unlock()

	return string(port), nil
}

func (s *LocalLeaser) Count() (string, error) {
	s.browserLock.RLock()
	count := len(s.browsers)
	s.browserLock.RUnlock()
	return strconv.Itoa(count), nil
}

func (s *LocalLeaser) Return(port string) error {
	s.browserLock.Lock()
	defer s.browserLock.Unlock()

	if b, ok := s.browsers[port]; ok {
		if err := b.ExitProcess(); err != nil {
			return err
		}
		delete(s.browsers, port)
		return nil
	}

	return errors.New("not found")
}

func (s *LocalLeaser) Cleanup() (string, error) {
	if err := KillOldProcesses(); err != nil {
		return "", err
	}

	if err := RemoveTmpContents(s.tmp); err != nil {
		return "", err
	}
	return "ok", nil
}
