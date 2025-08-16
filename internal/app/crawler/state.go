package crawler

import "sync"

type State interface {
	IsVisited(url string) bool
	MarkVisited(url string)
}

type SafeMapState struct {
	visited map[string]struct{}
	mu      sync.Mutex
}

func NewSafeMapState() *SafeMapState {
	return &SafeMapState{
		visited: make(map[string]struct{}),
	}
}

func (s *SafeMapState) IsVisited(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.visited[url]
	return ok
}

func (s *SafeMapState) MarkVisited(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.visited[url] = struct{}{}
}
