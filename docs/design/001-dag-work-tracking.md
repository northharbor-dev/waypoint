# Design: DAG-Based Work Tracking

**Status:** Approved
**Author:** @bobmhong
**Created:** 2026-03-01
**Updated:** 2026-03-01

## Summary

Waypoint models project work as a directed acyclic graph (DAG) of work items with typed dependencies, role-based ownership, and atomic claiming. Work items are persisted in MongoDB behind a `Store` interface, while all graph algorithms (validation, readiness, critical path) execute in-memory in Go. The CLI provides a self-documenting interface (`--agent-info`) that enables AI coding agents to discover capabilities and participate in coordinated workflows without external documentation.

## Motivation

AI-assisted development teams run multiple specialized agents concurrently. Without coordination:

- Agents duplicate effort by picking up the same task
- Agents start work before prerequisites are complete
- Crashed agents leave tasks permanently stuck
- Plan changes (adding/removing work items) can create invalid dependency chains
- Agents need external documentation or detailed cursor rules to understand the workflow

Existing tools (GitHub Issues, Jira, markdown checklists) don't model dependencies as a graph, don't provide atomic claiming, and aren't designed for AI agent interaction.

## Goals

- Model work items and dependencies as a validated DAG
- Provide atomic task claiming to prevent concurrent double-pickup
- Support role-based filtering so agents see only relevant work
- Detect and prevent invalid DAG states (cycles, dangling deps, status inconsistencies)
- Support DAG evolution as project plans change
- Track timestamps (started, completed) and compute durations
- Generate visualizations (mermaid, PNG) and status reports
- Self-documenting CLI interface for AI agent onboarding

## Non-Goals

- Scheduling or timeline enforcement (phase targets are informational only)
- User authentication or authorization (trust-based; MongoDB credentials gate access)
- Real-time notifications or webhooks (future consideration)
- Web UI or dashboard (CLI and generated reports are the interface)

## Detailed Design

### Overview

```
Seed YAML → waypoint init → MongoDB (work_items, phases)
                                    ↓
                    waypoint next/start/done → atomic state transitions
                                    ↓
                    waypoint viz/report → generated markdown/PNG
```

The system has three layers:

1. **CLI** -- user/agent interface, command parsing, output formatting
2. **DAG engine** -- in-memory graph algorithms, validation, readiness checks
3. **Store** -- persistence abstraction with MongoDB implementation

### Component Changes

All components are new (greenfield project).

```
cmd/waypoint/main.go          Entry point
internal/cli/                  Cobra command definitions
internal/dag/                  Graph algorithms
  graph.go                     Build adjacency list, topological sort
  validate.go                  Cycle detection, dangling deps, consistency
  readiness.go                 Find items ready to start
  critical_path.go             Longest-path computation
internal/store/
  store.go                     Store interface
  mongo/mongo.go               MongoDB implementation
internal/models/
  workitem.go                  WorkItem, Phase, Status types
internal/viz/
  mermaid.go                   DAG → mermaid text
  renderer.go                  Renderer interface
  local.go                     chromedp PNG rendering
  cloud.go                     HTTP-based PNG rendering
```

### Data Model

**WorkItem:**

```go
type Status string

const (
    StatusNotStarted Status = "not_started"
    StatusInProgress Status = "in_progress"
    StatusDone       Status = "done"
    StatusBlocked    Status = "blocked"
)

type WorkItem struct {
    ID              string    `bson:"_id" yaml:"id"`
    Title           string    `bson:"title" yaml:"title"`
    Phase           int       `bson:"phase" yaml:"phase"`
    Owner           string    `bson:"owner" yaml:"owner"`
    Role            string    `bson:"role" yaml:"role"`
    Status          Status    `bson:"status"`
    Dependencies    []string  `bson:"dependencies" yaml:"dependencies"`
    ClaimedBy       *string   `bson:"claimed_by,omitempty"`
    StartedAt       *time.Time `bson:"started_at,omitempty"`
    CompletedAt     *time.Time `bson:"completed_at,omitempty"`
    DurationSeconds *int64    `bson:"duration_seconds,omitempty"`
    BlockerNote     *string   `bson:"blocker_note,omitempty"`
    Project         string    `bson:"project"`
    UpdatedAt       time.Time `bson:"updated_at"`
}
```

**Phase:**

```go
type Phase struct {
    ID      string `bson:"_id"`
    Number  int    `bson:"number" yaml:"number"`
    Name    string `bson:"name" yaml:"name"`
    Target  string `bson:"target" yaml:"target"`
    Project string `bson:"project"`
}
```

**Seed file format:**

```yaml
project: breezy
phases:
  - number: 1
    name: Foundation
    target: "Apr 2026"

work_items:
  - id: WI-001
    title: "Review & Finalize Design Specs"
    phase: 1
    owner: human
    role: lead
    dependencies: []
  - id: WI-002
    title: "Backend Project Scaffolding"
    phase: 1
    owner: agent
    role: backend_dev
    dependencies: [WI-001]
```

### CLI Changes

**Lifecycle commands:**

| Command | Description |
|---------|-------------|
| `init <seed.yaml>` | Parse YAML, validate DAG, seed MongoDB. Fails if project already exists (use `sync`). |
| `sync <seed.yaml>` | Diff YAML against live state. Add new items, update modified, flag removals. Validate before applying. |
| `validate` | Run all integrity checks against live DAG. Exit code 0 if clean, 1 if issues found. |

**Mutation commands:**

| Command | Description |
|---------|-------------|
| `add <WI-ID> --title --phase --role [--deps]` | Add single item. Validates DAG before committing. |
| `remove <WI-ID>` | Remove item. Fails if dependents exist. `--force` strips deps. `--cascade` removes downstream. |

**Workflow commands:**

| Command | Description |
|---------|-------------|
| `next [--role R]` | List items where status=not_started and all deps are done. Filter by role. |
| `start <WI-ID>` | Atomic claim via findOneAndUpdate. Sets in_progress, claimed_by, started_at. Fails if already claimed or deps not met. |
| `done <WI-ID>` | Sets done, completed_at, computes duration_seconds. Clears claimed_by. |
| `block <WI-ID> -r "reason"` | Sets blocked, stores blocker_note. |
| `unblock <WI-ID>` | Clears blocker, resets to not_started. |
| `reset <WI-ID>` | Clears claimed_by/started_at, resets to not_started. For crash recovery. |

**Reporting commands:**

| Command | Description |
|---------|-------------|
| `status [--phase N]` | Table of all items with status, role, claimed_by, elapsed time. Flags stale items. |
| `status --output json` | Full DAG state as structured JSON -- includes all items with computed `ready` and `deps_met` fields, `ready_items` array, `critical_path`, and summary counts. Designed for programmatic inspection by agents. |
| `report [--output F]` | Full markdown report: summary, phase progress, completed, in progress, blocked, ready, stale, critical path. |
| `viz [--output F] [--format md\|png] [--renderer local\|cloud]` | Generate DAG diagram. |
| `critical-path` | Display longest dependency chain. |

**Global flags:**

| Flag | Description |
|------|-------------|
| `--agent-info` | Output structured onboarding guide for AI agents. Covers workflow, commands, rules, error handling. |

### Key Flows

**Atomic claim flow:**

```
1. Agent runs `waypoint start WI-009`
2. CLI calls store.ClaimWorkItem("breezy", "WI-009", "backend_dev")
3. MongoDB executes:
   findOneAndUpdate(
     {_id: "WI-009", status: "not_started"},
     {$set: {status: "in_progress", claimed_by: "backend_dev", started_at: now()}}
   )
4a. If document returned: claim succeeded, CLI prints confirmation
4b. If null returned: another agent claimed it, CLI prints error with guidance
```

**DAG sync flow:**

```
1. User updates seed/workitems.yaml (adds WI-027, removes WI-010, changes WI-019 deps)
2. User runs `waypoint sync seed/workitems.yaml`
3. CLI loads YAML and current MongoDB state
4. CLI builds in-memory DAG from merged state
5. Validation runs: cycle check, dangling deps, consistency
6a. If valid: apply changes (add new, update modified, flag/remove old)
6b. If invalid: reject with specific error, no changes applied
7. CLI prints summary of changes
```

**Readiness check flow:**

```
1. Agent runs `waypoint next --role backend_dev`
2. CLI loads all work items from MongoDB (single query)
3. DAG engine builds adjacency list
4. For each item where status=not_started and role=backend_dev:
   - Check all items in dependencies array
   - If ALL have status=done, item is ready
5. Return ready items sorted by ID
```

### Error Handling

| Scenario | Behavior |
|----------|----------|
| Claim conflict | `start` returns error with claimer identity. Agent runs `next` again. |
| Deps not met | `start` returns error listing unmet dependencies. |
| Task not found | `done`/`start` returns error. If task was removed, agent runs `next`. |
| Cycle in DAG | `validate`/`sync`/`add` rejects with cycle path. |
| Dangling dep | `validate`/`sync`/`add` rejects with missing ID. |
| Remove with dependents | `remove` fails listing dependent items. `--force` to override. |
| MongoDB unreachable | All commands fail with connection error and retry guidance. |

## Alternatives Considered

### Alternative 1: Markdown file as state store

Work items tracked in a structured markdown file, edited via StrReplace.

Rejected because:
- Concurrent agents editing the same file creates merge conflicts
- Agents on different git branches see different state
- No atomic claiming mechanism
- Graph queries require parsing markdown

### Alternative 2: Neo4j graph database

Native graph database with Cypher query language.

Rejected because:
- Neo4j Aura Free auto-pauses after 72 hours of inactivity
- Overkill for <100 work items
- Additional infrastructure to manage
- Graph algorithms at this scale run trivially fast in-memory in Go
- MongoDB was already provisioned for the primary project

### Alternative 3: SQLite embedded database

Zero-infrastructure embedded database.

Rejected because:
- Filesystem-bound; agents on different machines can't share state
- User specifically wanted a hosted solution to avoid local databases
- Noted as a future Store backend for solo/local use cases

## Security Considerations

- MongoDB credentials stored in environment variables or `.env` file (gitignored)
- No authentication within Waypoint itself; access is gated by MongoDB credentials
- The `reset` and `remove --force` commands should be restricted to human operators (enforced by convention, documented in `--agent-info`)
- No sensitive data stored in work items (titles, roles, statuses only)

## Performance Considerations

- All work items loaded in a single MongoDB query (~26-100 documents per project)
- Graph algorithms run in-memory; sub-millisecond for graphs under 1000 nodes
- `findOneAndUpdate` for atomic claims adds no meaningful latency
- PNG rendering via chromedp takes 1-3 seconds (headless Chrome startup); cloud rendering depends on network latency

## Testing Strategy

- **DAG engine unit tests:** Cycle detection with known cyclic/acyclic graphs, readiness with various dependency states, critical path with known longest paths
- **Store integration tests:** MongoDB operations with a test database, atomic claim race simulation
- **CLI integration tests:** Full command flows (init → next → start → done), error paths (claim conflict, invalid DAG)
- **Visualization tests:** Mermaid output format validation, PNG rendering smoke test

## Migration / Rollout

Greenfield project -- no migration needed. Rollout to consuming projects (e.g., Breezy) involves:

1. Install waypoint binary
2. Create `seed/workitems.yaml` with project work breakdown
3. Add `.cursor/rules/workplan-tracking.mdc` with agent protocol
4. Optionally add `Taskfile.yml` work tasks for human convenience
5. Run `waypoint init seed/workitems.yaml`

## Open Questions

- [ ] Should `--agent-info` output be versioned so agents can detect CLI version changes?
- [ ] Should `sync` support dry-run mode (`--dry-run`) to preview changes without applying?
- [ ] Should work items support arbitrary metadata/tags for future extensibility?

## References

- [Architecture](../architecture.md)
- [Roadmap](../roadmap.md)
- [MongoDB findOneAndUpdate](https://www.mongodb.com/docs/manual/reference/method/db.collection.findOneAndUpdate/)
- [Cobra CLI framework](https://cobra.dev/)
- [goreleaser](https://goreleaser.com/)
