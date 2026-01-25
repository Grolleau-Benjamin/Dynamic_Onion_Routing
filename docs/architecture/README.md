# Architecture Documentation

This directory contains detailed architectural documentation for the DOR client implementations.

## Client Implementations

### [CLI Client Architecture](./client_cli.md)

**Synchronous, Non-Interactive Mode**

The CLI client provides a simple, blocking interface for sending messages through the onion routing network. Key topics:

- Producer-Consumer pattern with channels
- Blocking execution model (Execute -> Close workflow)
- Event-driven logging architecture
- Graceful shutdown handling
- Thread synchronization patterns

### [TUI Client Architecture](./client_tui.md)

**Asynchronous, Interactive Mode**

The TUI client provides a responsive terminal user interface built with Bubble Tea framework. Key topics:

- Bubble Tea Model-View-Update (MVU) pattern
- Asynchronous command pattern for background tasks
- Recursive event listening loop
- UI responsiveness at 60fps
- Integration with concurrent client operations

## Design Patterns

Both implementations share these core patterns:

| Pattern                | Purpose                                |
| ---------------------- | -------------------------------------- |
| **Producer-Consumer**  | Decouples network I/O from UI updates  |
| **Event-Driven**       | Loose coupling between components      |
| **Channel-Based Sync** | Safe concurrent access to shared state |
| **Graceful Shutdown**  | Clean termination of goroutines        |

## Choosing Between CLI and TUI

- **Use CLI** when you need: Simple output, scriptable behavior, or background execution
- **Use TUI** when you need: Interactive control, real-time feedback, or visual status monitoring
