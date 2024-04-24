package client

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/jdbann/sosh/store"
	"github.com/jdbann/sosh/ui"
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

	user      store.User
	publicKey ssh.PublicKey

	screen tea.Model

	lg          *lipgloss.Renderer
	bannerStyle lipgloss.Style
	quitStyle   lipgloss.Style
	store       Store
}

type Params struct {
	Width     int
	Height    int
	PublicKey ssh.PublicKey
	Store     Store
	Lipgloss  *lipgloss.Renderer
}

func NewModel(params Params) Model {
	m := Model{
		width:       params.Width,
		height:      params.Height,
		publicKey:   params.PublicKey,
		lg:          params.Lipgloss,
		bannerStyle: lipgloss.Style{},
		quitStyle:   lipgloss.Style{},
		store:       params.Store,
	}

	m.bannerStyle = m.lg.NewStyle().Foreground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("10")).Padding(1, 5)
	m.quitStyle = m.lg.NewStyle().Foreground(lipgloss.Color("8"))

	return m
}

func (m Model) Init() tea.Cmd {
	return getUser(m.store, m.publicKey)
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
	case mustLogInMsg:
		m.screen = signup.NewModel(signup.Params{
			PublicKey: m.publicKey,
			Store:     m.store,
		})
		cmds = append(cmds, m.screen.Init())
	case ui.LoggedInMsg:
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

func getUser(st Store, key ssh.PublicKey) tea.Cmd {
	return func() tea.Msg {
		user, err := st.GetUser(key)
		if err != nil {
			if errors.Is(err, store.ErrUnknownUser) {
				return mustLogInMsg{}
			}
			return errMsg(err)
		}

		return ui.LoggedInMsg{
			User: user,
		}
	}
}

type (
	mustLogInMsg struct{}
	errMsg       error
)
