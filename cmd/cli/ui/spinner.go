package ui

import "github.com/pterm/pterm"

func HandleSpinnerStop(spinner *pterm.SpinnerPrinter) {
	if err := spinner.Stop(); err != nil {
		ShowError("Error stopping spinner: %v", err)
	}
}
