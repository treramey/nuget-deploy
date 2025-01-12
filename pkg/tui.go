package tui

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// StartTea the entry point for the UI. Initializes the model.
func StartTea(pr project.GormRepository, er entry.GormRepository) error {
	if f, err := tea.LogToFile("debug.log", "help"); err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	constants.Pr = &pr
	constants.Er = &er

	m, _ := InitProject() // TODO: can we acknowledge this error
	constants.P = tea.NewProgram(m, tea.WithAltScreen())
	if err := constants.P.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}
