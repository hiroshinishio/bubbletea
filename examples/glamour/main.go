package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const content = `
# Today’s Menu

## Appetizers

| Name        | Price | Notes                           |
| ---         | ---   | ---                             |
| Tsukemono   | $2    | Just an appetizer               |
| Tomato Soup | $4    | Made with San Marzano tomatoes  |
| Okonomiyaki | $4    | Takes a few minutes to make     |
| Curry       | $3    | We can add squash if you’d like |

## Seasonal Dishes

| Name                 | Price | Notes              |
| ---                  | ---   | ---                |
| Steamed bitter melon | $2    | Not so bitter      |
| Takoyaki             | $3    | Fun to eat         |
| Winter squash        | $3    | Today it's pumpkin |

## Desserts

| Name         | Price | Notes                 |
| ---          | ---   | ---                   |
| Dorayaki     | $4    | Looks good on rabbits |
| Banana Split | $5    | A classic             |
| Cream Puff   | $3    | Pretty creamy!        |

All our dishes are made in-house by Karen, our chef. Most of our ingredients
are from our garden or the fish market down the street.

Some famous people that have eaten here lately:

* [x] René Redzepi
* [x] David Chang
* [ ] Jiro Ono (maybe some day)

Bon appétit!
`

type example struct {
	viewport  viewport.Model
	helpStyle lipgloss.Style
}

func newExample(ctx tea.Context) (example, error) {
	const width = 78

	vp := viewport.New(ctx, width, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return example{}, err
	}

	str, err := renderer.Render(content)
	if err != nil {
		return example{}, err
	}

	vp.SetContent(str)

	return example{
		viewport: vp,
	}, nil
}

func (e example) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	var err error
	e, err = newExample(ctx)
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}
	e.helpStyle = ctx.NewStyle().Foreground(lipgloss.Color("241"))
	return e, nil
}

func (e example) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return e, tea.Quit
		default:
			var cmd tea.Cmd
			e.viewport, cmd = e.viewport.Update(ctx, msg)
			return e, cmd
		}
	default:
		return e, nil
	}
}

func (e example) View(ctx tea.Context) string {
	return e.viewport.View(ctx) + e.helpView()
}

func (e example) helpView() string {
	return e.helpStyle.Render("\n  ↑/↓: Navigate • q: Quit\n")
}

func main() {
	if _, err := tea.NewProgram(example{}).Run(); err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		os.Exit(1)
	}
}
