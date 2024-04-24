package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/jdbann/sosh/store"
	"github.com/jdbann/sosh/ui/feed"
)

const (
	host = "localhost"
	port = "23234"

	banner = "We're not ready to get sosh(al) yet..."
)

// connect with: ssh -p 23234 localhost

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(soshHandler),
			activeterm.Middleware(),
			logging.StructuredMiddleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

var globalStore = &store.MemoryStore{}

func soshHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := sess.Pty()

	renderer := bubbletea.MakeRenderer(sess)
	bannerStyle := renderer.NewStyle().Foreground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("10")).Padding(1, 5)
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	u, err := globalStore.GetUser(sess.PublicKey())
	if err != nil && !errors.Is(err, store.ErrUnknownUser) {
		log.Error("Could not get user", "error", err)
		return nil, nil
	}

	m := clientModel{
		width:  pty.Window.Width,
		height: pty.Window.Height,

		user: u,

		bannerStyle: bannerStyle,
		quitStyle:   quitStyle,
		lg:          renderer,
	}

	if u.Key == nil {
		m.screen = newSignupModel(globalStore, sess.PublicKey())
	} else {
		m.screen = feed.NewModel(feed.Params{
			Store:    globalStore,
			Username: u.Name,
			Lipgloss: renderer,
		})
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type clientModel struct {
	width  int
	height int

	user store.User

	screen tea.Model

	lg          *lipgloss.Renderer
	bannerStyle lipgloss.Style
	quitStyle   lipgloss.Style
}

func (m clientModel) Init() tea.Cmd {
	if m.screen != nil {
		return m.screen.Init()
	}

	return nil
}

func (m clientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case registeredMsg:
		m.user = msg.user
		m.screen = feed.NewModel(feed.Params{
			Store:    globalStore,
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

func (m clientModel) View() string {
	if m.screen != nil {
		return m.screen.View()
	}

	return lipgloss.JoinVertical(lipgloss.Right,
		lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.bannerStyle.Render(strings.Join([]string{banner, m.user.Name}, " "))),
		m.quitStyle.Render("Press q to quit"),
	)
}
