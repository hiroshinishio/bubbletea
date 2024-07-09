package main

// An example illustrating how to pipe in data to a Bubble Tea application.
// More so, this serves as proof that Bubble Tea will automatically listen for
// keystrokes when input is not a TTY, such as when data is piped or redirected
// in.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			fmt.Println("Error getting input:", err)
			os.Exit(1)
		}
	}

	model := model{initialValue: strings.TrimSpace(b.String())}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Couldn't start program:", err)
		os.Exit(1)
	}
}

type model struct {
	initialValue string
	userInput    textinput.Model
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	i := textinput.New(ctx)
	i.Prompt = ""
	i.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	i.Width = 48
	i.SetValue(m.initialValue)
	i.CursorEnd()
	i.Focus()

	m.userInput = i

	return m, textinput.Blink
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "esc", "enter":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.userInput, cmd = m.userInput.Update(ctx, msg)
	return m, cmd
}

func (m model) View(ctx tea.Context) string {
	return fmt.Sprintf(
		"\nYou piped in: %s\n\nPress ^C to exit",
		m.userInput.View(ctx),
	)
}
