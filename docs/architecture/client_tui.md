# TUI Architecture (Bubble Tea)

This document details the architecture of the interactive Terminal User Interface (TUI) mode. The TUI is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, strictly following the **Model-View-Update (MVU)** pattern.

**Current Reference Commit:** [9e112d0](https://github.com/Grolleau-Benjamin/Dynamic_Onion_Routing/commit/9e112d0a9c988ca94b99bea9663b4f707a89d374)

## 1. Overview

The TUI must remain responsive (60fps) while processing asynchronous background events from the Core Client. We achieve this using the **Recursive Command Pattern**.

### Key Concepts
- **Direct Command Integration:** Instead of a separate "bridge" goroutine, we use a `tea.Cmd` function (`waitForEvent`) that blocks on the channel read.
- **Bubble Tea Concurrency:** The Bubble Tea runtime executes `tea.Cmd` functions in its own managed goroutines.
- **Recursive Loop:** When an event is received in `Update()`, the model immediately issues a *new* `waitForEvent` command to listen for the next one. This creates an infinite listening loop without blocking the UI thread.

## 2. File Structure (MVU)

- **`model.go`**: Entry point. `Init()` kicks off the client simulation and the first listener command.
- **`state.go`**: Pure data structure holding the `Client` reference and UI logs.
- **`update.go`**: Handles `KeyMsg` (User input) and `client.Event` (Network events).
- **`view.go`**: Renders the state string using `strings.Builder`.

## 3. Lifecycle & Event Loop (Sequence)

The following diagram illustrates how the `waitForEvent` command acts as the glue between the Go Channel and the Bubble Tea Update loop.

```mermaid
sequenceDiagram
    autonumber
    participant TEA as BubbleTea Runtime
    participant Update as Update() (Main Thread)
    participant Cmd as Cmd: waitForEvent
    participant Chan as c.Events (Channel)
    participant Core as Client Core

    Note over TEA, Core: Phase 1: Initialization

    TEA->>Core: Init: client.Simulate()
    activate Core
    TEA->>Cmd: Init: Schedule waitForEvent()
    activate Cmd

    Note over TEA, Core: Phase 2: The Loop

    Core->>Chan: Push Event (Log)
    Cmd->>Chan: Read Event (Unblock)
    Chan-->>Cmd: Event Payload

    Cmd-->>TEA: Return Msg (client.Event)
    deactivate Cmd

    TEA->>Update: Call Update(msg)
    Update->>Update: Append Log to State

    Note right of Update: RECURSION START
    Update-->>TEA: Return Cmd: waitForEvent()
    TEA->>Cmd: Schedule waitForEvent()
    activate Cmd

    Note over TEA, Core: Phase 3: Termination

    Note left of TEA: User presses 'q'
    TEA->>Update: KeyMsg('q')
    Update->>Core: client.Close()
    Core->>Chan: close()
    deactivate Core

    Update-->>TEA: Quit
    deactivate Cmd
```

## The "Recursive Command" Data Flow
This flowchart shows the circular nature of the event handling. Notice how Update() is responsible for re-arming the listener.

```mermaid
flowchart TB
    subgraph Core_Layer
        ClientLogic["Client Logic"]
        ClientChan[("Channel: c.Events")]
    end

    subgraph BubbleTea_Loop ["The Elm Architecture"]
        Init(("Init()"))
        Update["Update()"]
        View["View()"]

        subgraph Managed_Commands ["Runtime Managed Goroutines"]
            WaitCmd["Cmd: waitForEvent(ch)"]
        end

        subgraph Messages ["Messages"]
            MsgEv["Msg: client.Event"]
            MsgKey["Msg: KeyMsg (q)"]
        end
    end

    Init -- "1. Start Simulation" --> ClientLogic
    Init -- "2. Initial Wait" --> WaitCmd

    ClientLogic -- "3. Emit" --> ClientChan
    ClientChan -- "4. Read (Block)" --> WaitCmd
    WaitCmd -- "5. Return" --> MsgEv

    MsgEv --> Update
    Update -- "6. Modify State" --> View
    Update -- "7. RECURSE: New Wait" --> WaitCmd

    MsgKey --> Update
    Update -- "8. Close & Quit" --> ClientLogic

    linkStyle 3,4 stroke:orange;
    linkStyle 7 stroke:green,stroke-width:2px;
```
