# sosh - social over SSH

A simple social media application. Access it via SSH. Use your public keys to authenticate. Get stuck talking to the kind of people that want to join a social network only accessible via SSH.

## Progress

- [x] Setup a basic [bubbletea] app that can be reached via SSH using [wish]
- [ ] Add user registration and authentication via [public key auth] (persistened in memory)
- [ ] Add posting and reading (persisted in memory)
- [ ] Add a persistent [SQLite] store
- [ ] Host this somewhere

## Development

Running locally is simple:

```sh
go run .
```

Testing locally is simple too:

```sh
ssh -p 23234 localhost
```

This [pro tip](https://github.com/charmbracelet/wish?tab=readme-ov-file#pro-tip) can be helpful for development. And there is a dangerous ssh option that I do not advise using: `-o StrictHostKeyChecking=accept-new`.

[bubbletea]: https://github.com/charmbracelet/bubbletea
[wish]: https://github.com/charmbracelet/wish
[public key auth]: https://pkg.go.dev/github.com/charmbracelet/wish#WithPublicKeyAuth
[SQLite]: https://www.sqlite.org/index.html
