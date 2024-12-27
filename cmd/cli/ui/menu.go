package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// MenuOption représente une option dans un menu interactif
type MenuOption struct {
	Label    string
	Handler  func() error
	Position int
}

// MenuBuilder aide à construire les menus interactifs
type MenuBuilder struct {
	options []MenuOption
}

// NewMenuBuilder crée un nouveau builder de menu
func NewMenuBuilder() *MenuBuilder {
	return &MenuBuilder{
		options: make([]MenuOption, 0),
	}
}

// AddOption ajoute une option au menu
func (m *MenuBuilder) AddOption(label string, handler func() error) *MenuBuilder {
	m.options = append(m.options, MenuOption{
		Label:    label,
		Handler:  handler,
		Position: len(m.options) + 1,
	})
	return m
}

// Build crée et affiche le menu interactif
func (m *MenuBuilder) Build() error {
	options := make([]string, len(m.options))
	handlers := make(map[string]func() error, len(m.options))

	for i, opt := range m.options {
		key := fmt.Sprintf("%d. %s", opt.Position, opt.Label)
		options[i] = key
		handlers[key] = opt.Handler
	}

	selected, err := pterm.DefaultInteractiveSelect.
		WithMaxHeight(10).
		WithOptions(options).
		Show()
	if err != nil {
		return fmt.Errorf("error displaying menu: %w", err)
	}

	return handlers[selected]()
}

// AddStandardOptions ajoute les options standard (Skip, Exit) au menu
func (m *MenuBuilder) AddStandardOptions(skipHandler, exitHandler func() error) *MenuBuilder {
	return m.
		AddOption("Skip", skipHandler).
		AddOption("Exit", exitHandler)
}
