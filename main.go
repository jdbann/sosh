package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
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

func soshHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := sess.Pty()

	renderer := bubbletea.MakeRenderer(sess)
	bannerStyle := renderer.NewStyle().Foreground(lipgloss.Color("10")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("10")).Padding(1, 5)
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	m := clientModel{
		width:  pty.Window.Width,
		height: pty.Window.Height,

		bannerStyle: bannerStyle,
		quitStyle:   quitStyle,
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type clientModel struct {
	width  int
	height int

	bannerStyle lipgloss.Style
	quitStyle   lipgloss.Style
}

func (m clientModel) Init() tea.Cmd {
	return nil
}

func (m clientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m clientModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Right,
		lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.bannerStyle.Render(banner)),
		m.quitStyle.Render("Press q to quit"),
	)
}
