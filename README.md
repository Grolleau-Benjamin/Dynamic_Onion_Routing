# DOR - Dynamic Onion Routing

## Documentation
- [Client Cli Architecture](./docs/architecture/client_cli.md)
- [Client TUI Architecture](./docs/architecture/client_tui.md)

## Server Daemon
```shell
go run cmd/dord/main.go
```

## Client
```shell
go run cmd/dorc/main.go \
  --dest "[cafe::1]:8080" \
  --onion-path "[abcd::1]:62503,127.0.0.1:62503|[baba::13]:1418" \
  --payload "Some dada" \
  --tui # Optional -> Enable Bubbletea TUI
```

## Plugins
In addition to the demonstrator, this protocol also provides a Wireshark Lua plugin to observe DOR protocol exchanges.

Copy the plugin to your Wireshark plugins directory:

```shell
# Linux / Macos sample
cp ./tools/wireshark/dor.lua ~/.config/wireshark/plugins/
```

> [!NOTE]
> The plugin does not decrypt ciphertext - it only displays the protocol structure visible on the network.
