# UpsilonAPI: Tactical RPG Engine Bridge

**UpsilonAPI** ("the Bridge") is a high-performance, Go-based HTTP JSON API embedding the **UpsilonBattle** engine (`BattleArena`). It handles all game mechanics, from board generation and initiative calculations to complex combat resolutions, operating as an isolated, stateless logic engine.

Built for scalability and precision, it uses an actor-model architecture to manage multiple concurrent skirmishes (Arenas) while providing a standardized JSON interface for orchestration by **UpsilonHub**, the Go platform gateway that owns auth, matchmaking, economy, and admin for the Upsilon platform.

## Key Responsibilities

- **Stateless Game Logic:** Calculates HP reduction, defense mitigation, and attribute progression impacts.
- **Wait-Time Engine:** Governs the non-linear initiative system and manages the sequence of character actions.
- **Board Orchestration:** Generates tactical grids with procedural obstacles and manages real-time entity positioning.
- **Action Proxying:** Translates high-level HTTP commands (Move, Attack, Pass) into deterministic engine operations.
- **Real-time Telemetry:** Broadcasts game state updates to registered callback URLs for real-time visualization.

## Getting Started

### Prerequisites
- Go 1.25+

### Installation & Run

#### Development (Run from Source)
To start the engine directly from source:

```bash
go run main.go
```

#### Production Build
To compile a binary for distribution or deployment:

```bash
go build -o bin/upsilonapi .
./bin/upsilonapi
```

The server will start on `:8081` by default.

## Project Structure

- **[/api](api)**: Defines the core data structures and standard network envelopes for request/response payloads.
- **[/bridge](bridge)**: The transition layer between the HTTP handlers and the underlying actor-based engine logic (embeds `upsilonbattle`'s `BattleArena`).
- **[/handler](handler)**: Contains the Gin-gonic HTTP handlers for the `/v1` api group.
- **[/stdmessage](stdmessage)**: Formatting utilities for standard system-wide logging and message envelopes.
- **[main.go](main.go)**: Application entry point and router initialization.

## Integration Architecture

UpsilonAPI occupies the **Architecture/Implementation** boundary of the system.

### Orchestration by UpsilonHub
[UpsilonHub](https://github.com/ecumeurs/upsilonhub) is the platform gateway: it owns auth/identity, matchmaking, economy, admin, and the SSE realtime layer. Its `internal/games/battle` module talks to UpsilonAPI over HTTP via a typed `engineclient` (implementing the hub's `battle.Engine` contract), calling the `/v1/arena/*` endpoints on port `:8081` to start arenas, proxy actions, and resurrect state. The engine posts webhook events back to a `callback_url` supplied at arena start, which the hub consumes to update live state and fan it out over SSE.

### Verification via UpsilonCLI
The [UpsilonCLI](https://github.com/ecumeurs/upsiloncli) provides a direct line of sight into the API. It can be used to simulate full combat sequences, verify response payloads, and monitor real-time board updates.

## ATD Traceability

This module is strictly governed by the **Atomic Traceable Documentation (ATD)** framework. Key specifications include:

- **[[module_upsilonapi]]**: Architectural blueprint for the Go-based engine bridge.
- **[[api_go_battle_engine]]**: External JSON boundary for the engine (base path `/v1`, port `8081`).
- **[[api_go_webhook_callback]]**: Asynchronous event delivery to the calling gateway's `callback_url`.
- **[[api_standard_envelope]]**: Standardized format for all network communication.

---
*Note: This engine does not handle persistent player state, auth, or matchmaking — those are owned by UpsilonHub. UpsilonAPI is a stateless (in-memory per-arena) logic engine.*
