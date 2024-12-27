package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// PromptText demande à l'utilisateur de saisir du texte
func PromptText(message, defaultValue string) (string, error) {
	result, err := pterm.DefaultInteractiveTextInput.
		WithDefaultValue(defaultValue).
		Show(message)

	if err != nil {
		return "", fmt.Errorf("error getting user input: %w", err)
	}

	return result, nil
}

// ShowInfo affiche un message d'information
func ShowInfo(message string, args ...interface{}) {
	pterm.Info.Printfln(message, args...)
}

// ShowSuccess affiche un message de succès
func ShowSuccess(message string, args ...interface{}) {
	pterm.Success.Printfln(message, args...)
}

// ShowError affiche un message d'erreur
func ShowError(message string, args ...interface{}) {
	pterm.Error.Printfln(message, args...)
}
