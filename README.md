# UpsilonAPI: Tactical RPG Engine

**UpsilonAPI** is the high-performance, Go-based calculating brain behind the **UpsilonBattle** tactical RPG. It handles all game mechanics, from board generation and initiative calculations to complex combat resolutions, operating as an isolated, stateless logic engine.

Built for scalability and precision, it uses an actor-model architecture to manage multiple concurrent skirmishes (Arenas) while providing a standardized JSON interface for orchestration by external gateways (like `BattleUI`).

## Key Responsibilities

- **Stateless Game Logic:** Calculates HP reduction, defense mitigation, and attribute progression impacts.
- **Wait-Time Engine:** Governs the non-linear initiative system and manages the sequence of character actions.
- **Board Orchestration:** Generates tactical grids with procedural obstacles and manages real-time entity positioning.
- **Action Proxying:** Translates high-level HTTP commands (Move, Attack, Pass) into deterministic engine operations.
- **Real-time Telemetry:** Broadcasts game state updates to registered callback URLs for real-time visualization.

## Getting Started

### Prerequisites
- Go 1.22+

### Installation & Run
To start the engine locally:

```bash
go run main.go
```
The server will start on `:8081` by default.

## Project Structure

- **[/api](file:///workspace/upsilonapi/api)**: Defines the core data structures and standard network envelopes for request/response payloads.
- **[/bridge](file:///workspace/upsilonapi/bridge)**: The transition layer between the HTTP handlers and the underlying actor-based engine logic.
- **[/handler](file:///workspace/upsilonapi/handler)**: Contains the Gin-gonic HTTP handlers for the `/internal` and `/v1` api groups.
- **[/stdmessage](file:///workspace/upsilonapi/stdmessage)**: Formatting utilities for standard system-wide logging and message envelopes.
- **[main.go](file:///workspace/upsilonapi/main.go)**: Application entry point and router initialization.

## Integration Architecture

UpsilonAPI occupies the **Architecture/Implementation** boundary of the system.

### Orchestration by BattleUI
The [BattleUI](file:///workspace/battleui) (Laravel) acts as the gateway. It manages player sessions and matchmaking, then "hands off" the combat logic to UpsilonAPI by calling the `/internal/arena/start` endpoint.

### Verification via UpsilonCLI
The [UpsilonCLI](file:///workspace/upsiloncli) provides a direct line of sight into the API. It can be used to simulate full combat sequences, verify response payloads, and monitor real-time board updates via WebSockets.

## ATD Traceability

This module is strictly governed by the **Atomic Traceable Documentation (ATD)** framework. Key specifications include:

- **[[module_upsilonapi]]**: Architectural blueprint for the Go-based engine bridge.
- **[[rule_initiative_delay]]**: Specification for the non-linear ticker-based turn order.
- **[[rule_team_mechanics]]**: logic for ally/enemy recognition and friendly-fire prevention.
- **[[api_standard_envelope]]**: Standardized format for all network communication.

---
*Note: This engine does not handle persistent player state or matchmaking. Use the [BattleUI](file:///workspace/battleui) for session management.*
