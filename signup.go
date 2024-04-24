package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/jdbann/sosh/store"
)

const (
	nameKey = "name"
)

type UserStore interface {
	GetUser(ssh.PublicKey) (store.User, error)
	AddUser(ssh.PublicKey, string) error
}

type signupModel struct {
	key   ssh.PublicKey
	form  *huh.Form
	err   error
	store UserStore
}

func newSignupModel(store UserStore, key ssh.PublicKey) signupModel {
	m := signupModel{
		key:   key,
		store: store,
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

func (m signupModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m signupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m signupModel) View() string {
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

		return registeredMsg{
			user: u,
		}
	}
}

type (
	registeredMsg struct {
		user store.User
	}
	registrationErrMsg error
)
