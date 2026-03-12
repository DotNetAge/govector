package api

import (
	"github.com/DotNetAge/govector/core"
)

// GetCollection safely retrieves a collection for testing purposes.
func (s *Server) GetCollection(name string) *core.Collection {
	s.serverMu.Lock()
	defer s.serverMu.Unlock()
	return s.collections[name]
}

func (s *Server) GetCollectionsMapSize() int {
	s.serverMu.Lock()
	defer s.serverMu.Unlock()
	return len(s.collections)
}
