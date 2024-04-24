package signup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/jdbann/sosh/store"
	"github.com/jdbann/sosh/ui"
)

const (
	nameKey = "name"
)

type UserStore interface {
	GetUser(ssh.PublicKey) (store.User, error)
	AddUser(ssh.PublicKey, string) error
}

type Model struct {
	key   ssh.PublicKey
	form  *huh.Form
	err   error
	store UserStore
}

type Params struct {
	PublicKey ssh.PublicKey
	Store     UserStore
}

func NewModel(params Params) Model {
	m := Model{
		key:   params.PublicKey,
		store: params.Store,
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key(nameKey).
				Title("Name"),
		),
	).
		WithShowHelp(true)

	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case registrationErrMsg:
		m.err = msg
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		cmds = append(cmds, registerCmd(m.store, m.key, m.form.GetString(nameKey)))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return m.form.View()
}

func registerCmd(store UserStore, key ssh.PublicKey, name string) tea.Cmd {
	return func() tea.Msg {
		if err := store.AddUser(key, name); err != nil {
			log.Error("Registering user", "error", err)
			return registrationErrMsg(err)
		}

		u, err := store.GetUser(key)
		if err != nil {
			log.Error("Getting user", "error", err)
			return registrationErrMsg(err)
		}

		return ui.LoggedInMsg{
			User: u,
		}
	}
}

type (
	registrationErrMsg error
)
