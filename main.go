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
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/jdbann/sosh/store"
	"github.com/jdbann/sosh/ui/client"
	"github.com/jdbann/sosh/ui/feed"
	"github.com/jdbann/sosh/ui/signup"
)

const (
	host = "localhost"
	port = "23234"
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

	u, err := globalStore.GetUser(sess.PublicKey())
	if err != nil && !errors.Is(err, store.ErrUnknownUser) {
		log.Error("Could not get user", "error", err)
		return nil, nil
	}

	params := client.Params{
		Width:    pty.Window.Width,
		Height:   pty.Window.Height,
		User:     u,
		Store:    globalStore,
		Lipgloss: renderer,
	}

	if u.Key == nil {
		params.Screen = signup.NewModel(signup.Params{
			PublicKey: sess.PublicKey(),
			Store:     globalStore,
		})
	} else {
		params.Screen = feed.NewModel(feed.Params{
			Store:    globalStore,
			Username: u.Name,
			Lipgloss: renderer,
		})
	}

	m := client.NewModel(params)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
