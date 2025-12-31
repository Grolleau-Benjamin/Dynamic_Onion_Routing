# Dynamic Onion Routing

## Documentation

- [Client Cli Architecture](./docs/architecture/client_cli.md)
- [Client TUI Architecture](./docs/architecture/client_tui.md)

## Server Daemon
```shell
go run cmd/dord/main.go
```

## Client Daemon
```shell
go run cmd/dorc/main.go --dest "[cafe::1]:8080" --onion-path "[abcd::1]:62503,127.0.0.1:62503|[baba::13]:1418" --payload "test" --log-level debug
```
