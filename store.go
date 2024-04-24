package main

import (
	"errors"

	"github.com/charmbracelet/ssh"
)

var errUnknownUser = errors.New("unknown user")

var users = []user{}

type user struct {
	name string
	key  []byte
}

func getUser(key ssh.PublicKey) (user, error) {
	for _, u := range users {
		parsed, err := ssh.ParsePublicKey([]byte(u.key))
		if err != nil {
			return user{}, err
		}

		if ssh.KeysEqual(key, parsed) {
			return u, nil
		}
	}
	return user{}, errUnknownUser
}

func addUser(key ssh.PublicKey, name string) error {
	keyBytes := key.Marshal()
	u := user{
		name: name,
		key:  keyBytes,
	}
	users = append(users, u)
	return nil
}
