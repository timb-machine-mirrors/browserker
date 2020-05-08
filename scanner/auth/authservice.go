package auth

import "gitlab.com/browserker/browserk"

type Service struct {
}

func New(cfg *browserk.Config) *Service {
	return &Service{}
}

func (s *Service) Init() error {
	return nil
}

func (s *Service) Login(c *browserk.Context) {

}
