package feed

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jdbann/sosh/store"
)

const (
	bodyKey = "body"
)

type PostsStore interface {
	GetPosts() ([]store.Post, error)
	AddPost(store.Post) error
}

type Model struct {
	posts    []store.Post
	store    PostsStore
	username string

	addForm *huh.Form

	bodyStyle   lipgloss.Style
	authorStyle lipgloss.Style
	metaStyle   lipgloss.Style
}

type Params struct {
	Store    PostsStore
	Username string
	Lipgloss *lipgloss.Renderer
}

func NewModel(params Params) Model {
	return Model{
		store:    params.Store,
		username: params.Username,

		bodyStyle:   params.Lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		authorStyle: params.Lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		metaStyle:   params.Lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

func (m Model) Init() tea.Cmd {
	return getPosts(m.store)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			if m.addForm != nil {
				break
			}
			m.addForm = huh.NewForm(
				huh.NewGroup(
					huh.NewText().Key(bodyKey),
				),
			)
			return m, m.addForm.Init()
		}
	case postsMsg:
		m.posts = msg
	}

	var cmds []tea.Cmd

	if m.addForm != nil {
		form, cmd := m.addForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.addForm = f
			cmds = append(cmds, cmd)
		}

		if m.addForm.State == huh.StateCompleted {
			cmds = append(cmds, addPost(m.store, store.Post{
				Author:      m.username,
				Body:        m.addForm.GetString(bodyKey),
				PublishedAt: time.Now(),
			}))
			m.addForm = nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var out strings.Builder

	for _, p := range m.posts {
		out.WriteString(m.bodyStyle.Render(p.Body))
		out.WriteString("\n")
		out.WriteString(m.authorStyle.Render(p.Author))
		out.WriteString(m.metaStyle.Render(" -", p.PublishedAt.Format(time.RFC3339Nano)))
		out.WriteString("\n\n")
	}

	if m.addForm != nil {
		out.WriteString(m.addForm.View())
	}

	return out.String()
}

func getPosts(store PostsStore) tea.Cmd {
	return func() tea.Msg {
		posts, err := store.GetPosts()
		if err != nil {
			log.Error("Getting posts", "error", err)
			return postsErr(err)
		}

		return postsMsg(posts)
	}
}

func addPost(store PostsStore, post store.Post) tea.Cmd {
	return func() tea.Msg {
		if err := store.AddPost(post); err != nil {
			log.Error("Adding post", "error", err)
			return postsErr(err)
		}

		return getPosts(store)()
	}
}

type (
	postsErr error
	postsMsg []store.Post
)
