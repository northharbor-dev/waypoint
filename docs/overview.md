# Waypoint Overview

Waypoint is a command-line work tracking tool designed for development teams that include AI coding agents alongside human engineers. It models work items and their dependencies as a directed acyclic graph (DAG), provides concurrency-safe task claiming, and offers a self-documenting interface that AI agents can discover and use autonomously.

## Key Capabilities

### Dependency-Aware Task Discovery

Work items form a DAG. The `next` command finds tasks whose dependencies are all complete and filters by role, so each agent or human sees only the work they can and should pick up.

### Atomic Task Claiming

When multiple agents share the same role, MongoDB's `findOneAndUpdate` ensures only one can claim a given task. If a claim fails, the agent is directed to find alternative work.

### DAG Validation and Evolution

The `validate` command checks for cycles, dangling dependencies, status inconsistencies, and invalid roles. The `sync` command reconciles an updated YAML seed file with live state -- adding new items, updating modified ones, and flagging removals for review -- all while preserving existing status and timestamps.

### Status Reporting and Visualization

Generate markdown status reports, mermaid DAG diagrams (markdown or PNG), and critical path analysis. Reports include completion timestamps, durations, and stale task detection.

### Agent Self-Onboarding

The `--agent-info` flag outputs a structured guide covering workflow, commands, rules, error handling, and examples. A project's cursor rule can simply say "run `waypoint --agent-info`" rather than encoding the full protocol.

### Crash Recovery

If an agent crashes mid-task, `waypoint status` flags stale items and `waypoint reset` releases them back to the pool. No heartbeats or leases -- human review at this scale is sufficient.

## Target Users

- **Solo developers** using AI coding agents for parallel development tracks
- **Small teams** coordinating work across specialized agents (API, UI, backend, security, testing)
- **Organizations** adopting AI-assisted development workflows

## Use Cases

1. **Multi-agent project execution** -- seed a work breakdown, let agents discover and claim tasks by role, track progress through completion
2. **Project status reporting** -- generate markdown reports and DAG visualizations for stakeholders, Confluence, or GitHub
3. **Plan evolution** -- add, remove, or re-sequence work items as the project plan changes, with validation preventing invalid states

## Technology Stack

- **Language:** Go
- **CLI framework:** Cobra
- **Database:** MongoDB (via Store interface -- swappable)
- **Visualization:** Mermaid text generation, PNG via chromedp (local) or mermaid.ink (cloud)
- **Distribution:** Cross-platform binaries via goreleaser
