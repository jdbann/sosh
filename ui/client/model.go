package client

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jdbann/sosh/store"
	"github.com/jdbann/sosh/ui/feed"
	"github.com/jdbann/sosh/ui/signup"
)

const (
	banner = "We're not ready to get sosh(al) yet..."
)

type Store interface {
	feed.PostsStore
	signup.UserStore
}

type Model struct {
	width  int
	height int

	user store.User

	screen tea.Model

	lg          *lipgloss.Renderer
	bannerStyle lipgloss.Style
	quitStyle   lipgloss.Style
	store       Store
}

type Params struct {
	Width    int
	Height   int
	User     store.User
	Store    Store
	Lipgloss *lipgloss.Renderer
	Screen   tea.Model
}

func NewModel(params Params) Model {
	m := Model{
		width:  params.Width,
		height: params.Height,
		user:   params.User,
		screen: params.Screen,
		lg:     params.Lipgloss,
		store:  params.Store,
	}

	m.bannerStyle = m.lg.NewStyle().Foreground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("10")).Padding(1, 5)
	m.quitStyle = m.lg.NewStyle().Foreground(lipgloss.Color("8"))

	return m
}

func (m Model) Init() tea.Cmd {
	if m.screen != nil {
		return m.screen.Init()
	}

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case signup.RegisteredMsg:
		m.user = msg.User
		m.screen = feed.NewModel(feed.Params{
			Store:    m.store,
			Username: m.user.Name,
			Lipgloss: m.lg,
		})
		cmds = append(cmds, m.screen.Init())
	}

	if m.screen != nil {
		screen, cmd := m.screen.Update(msg)
		m.screen = screen
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.screen != nil {
		return m.screen.View()
	}

	return lipgloss.JoinVertical(lipgloss.Right,
		lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.bannerStyle.Render(strings.Join([]string{banner, m.user.Name}, " "))),
		m.quitStyle.Render("Press q to quit"),
	)
}
