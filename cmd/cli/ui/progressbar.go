package ui

import (
	"github.com/pterm/pterm"
)

func HandlePbStop(pb *pterm.ProgressbarPrinter) {
	if _, err := pb.Stop(); err != nil {
		ShowError("Error stopping progress bar: %v", err)
	}
}
