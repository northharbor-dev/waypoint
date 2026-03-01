# Waypoint

**Graph-based work tracking for AI-assisted development teams.**

Waypoint tracks work items and their dependencies as a directed acyclic graph (DAG), backed by MongoDB. It answers one question well: **"What should I work on next?"** -- for both humans and AI coding agents.

## Why Waypoint

When your development team includes specialized AI agents working concurrently, you need:

- **Dependency-aware task assignment** -- agents should only pick up work whose prerequisites are complete
- **Concurrency safety** -- two agents with the same role should never claim the same task
- **Crash recovery** -- if an agent fails mid-task, the work item shouldn't be stuck forever
- **DAG integrity** -- as the plan evolves, invalid dependency chains should be caught before they cause problems
- **Self-documenting interface** -- agents should be able to discover capabilities and onboard without external docs

Waypoint solves all of these with a single Go binary and a MongoDB collection.

## Quick Start

### Prerequisites

- MongoDB Atlas account (or any MongoDB instance)
- Go 1.22+ (for building from source) or download a [release](https://github.com/northharbor-dev/waypoint/releases)

### Install

```bash
go install github.com/northharbor-dev/waypoint/cmd/waypoint@latest
```

Or download a pre-built binary from [Releases](https://github.com/northharbor-dev/waypoint/releases).

### Configure

```bash
export WAYPOINT_MONGO_URI="mongodb+srv://..."
export WAYPOINT_MONGO_DATABASE="waypoint"
export WAYPOINT_PROJECT="my-project"
```

Or create a `.env` file (see [`.env.example`](.env.example)).

### Seed a Project

Define your work breakdown in YAML:

```yaml
project: my-project
phases:
  - number: 1
    name: Foundation
    target: "Q2 2026"

work_items:
  - id: WI-001
    title: "Define API contracts"
    phase: 1
    owner: human
    role: lead
    dependencies: []
  - id: WI-002
    title: "Implement API endpoints"
    phase: 1
    owner: agent
    role: api_dev
    dependencies: [WI-001]
```

```bash
waypoint init seed/workitems.yaml
```

### Work the Plan

```bash
waypoint next --role api_dev    # find ready tasks for your role
waypoint start WI-002           # claim a task (atomic, prevents double-pickup)
# ... do the work ...
waypoint done WI-002            # mark complete
```

## Commands

```
LIFECYCLE:
  waypoint init <seed.yaml>            Seed a new project from YAML
  waypoint sync <seed.yaml>            Reconcile updated YAML with live state
  waypoint validate                    Check DAG integrity

MUTATIONS:
  waypoint add <WI-ID> --title "..." --phase N --role <role> [--deps WI-X,WI-Y]
  waypoint remove <WI-ID>

WORKFLOW:
  waypoint next [--role <role>]        Show tasks ready to start
  waypoint start <WI-ID>               Claim task atomically
  waypoint done <WI-ID>                Mark task complete
  waypoint block <WI-ID> -r "reason"   Mark blocked
  waypoint unblock <WI-ID>             Clear blocker
  waypoint reset <WI-ID>               Release orphaned task (crash recovery)

REPORTING:
  waypoint status [--phase N]          Show all tasks with current state
              [--output json]          Full DAG state as JSON (for agent inspection)
  waypoint report [--output file]      Generate markdown status report
  waypoint viz [--output file] [--format md|png]
  waypoint critical-path               Show longest dependency chain

HELP:
  waypoint help [command]              Help for humans
  waypoint --agent-info                Self-contained guide for AI agents
```

## For AI Agents

Run `waypoint --agent-info` for a complete onboarding guide including workflow, rules, error handling, and examples. This is the canonical entry point for agents -- no external documentation needed.

## Architecture

- **Go binary** -- single executable, cross-platform (darwin/linux/windows, amd64/arm64)
- **MongoDB** -- persistence layer behind a `Store` interface (swappable backends)
- **In-memory DAG engine** -- graph algorithms run in Go, not in database queries
- **Atomic claims** -- MongoDB `findOneAndUpdate` prevents concurrent double-pickup
- **DAG validation** -- cycle detection, dangling dependency checks, status consistency

See [docs/architecture.md](docs/architecture.md) for details.

## Documentation

- [Getting Started](docs/getting-started.md) -- install, configure MongoDB Atlas, seed a project, and run your first workflow
- [Overview](docs/overview.md) -- what Waypoint is and who it's for
- [Architecture](docs/architecture.md) -- system design and data model
- [Roadmap](docs/roadmap.md) -- current and planned work
- [Design Documents](docs/design/) -- detailed technical designs

## About

Waypoint is an open-source project by [NorthHarbor Development](https://github.com/northharbor-dev), created by Bob Hong.

## License

Apache-2.0
