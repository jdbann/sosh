package store

import (
	"errors"

	"github.com/charmbracelet/ssh"
)

var ErrUnknownUser = errors.New("unknown user")

type MemoryStore struct {
	Posts []Post
	Users []User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) GetPosts() ([]Post, error) {
	return s.Posts, nil
}

func (s *MemoryStore) AddPost(post Post) error {
	s.Posts = append(s.Posts, post)
	return nil
}

func (s *MemoryStore) GetUser(key ssh.PublicKey) (User, error) {
	for _, u := range s.Users {
		parsed, err := ssh.ParsePublicKey(u.Key)
		if err != nil {
			return User{}, err
		}

		if ssh.KeysEqual(key, parsed) {
			return u, nil
		}
	}

	return User{}, ErrUnknownUser
}

func (s *MemoryStore) AddUser(key ssh.PublicKey, name string) error {
	keyBytes := key.Marshal()
	user := User{
		Name: name,
		Key:  keyBytes,
	}
	s.Users = append(s.Users, user)
	return nil
}
