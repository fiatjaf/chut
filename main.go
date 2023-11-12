package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	m       = initialModel()
	program = tea.NewProgram(m)
)

func main() {
	os.WriteFile("debug.log", []byte("=*=\n"), 0644)

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(5)
	}
	defer f.Close()

	go subscribeToMessages(m)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
