package handlers

import (
	"context"

	"github.com/nouuu/gonamer/conf"
)

// MediaHandler définit l'interface commune pour tous les handlers de médias
type MediaHandler interface {
	// Handle traite une suggestion de média
	Handle(ctx context.Context) error
}

// BaseHandler contient les éléments communs à tous les handlers
type BaseHandler struct {
	Config     conf.Config
	DryRun     bool
	QuickMode  bool
	MaxResults int
}

// MediaSuggestion reste l'interface commune pour les suggestions
type MediaSuggestion interface {
	GetOriginalFilename() string
	GenerateFilename(pattern string) string
	HasSingleSuggestion() bool
	RenameFileManually(ctx context.Context) error
}

// NewBaseHandler crée un nouveau BaseHandler avec la configuration donnée
func NewBaseHandler(config conf.Config) BaseHandler {
	return BaseHandler{
		Config:     config,
		DryRun:     config.DryRun,
		QuickMode:  config.QuickMode,
		MaxResults: config.MaxResults,
	}
}
