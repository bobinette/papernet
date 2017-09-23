package oauth

import (
	"sync"
)

type ProviderService struct {
	mu        sync.Locker
	providers []string
}

func NewProviderService() *ProviderService {
	return &ProviderService{
		mu:        &sync.Mutex{},
		providers: make([]string, 0),
	}
}

func (s *ProviderService) Register(provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.providers {
		if p == provider {
			return
		}
	}

	s.providers = append(s.providers, provider)
}

func (s *ProviderService) Providers() []string {
	providers := make([]string, len(s.providers))
	s.mu.Lock()
	copy(providers, s.providers)
	s.mu.Unlock()
	return providers
}
