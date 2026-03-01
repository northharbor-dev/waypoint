# Roadmap

This document outlines the current and planned work for Waypoint.

## Current Phase: Core CLI

Building the foundational CLI with all essential commands and MongoDB persistence.

### In Progress

- [ ] Go project scaffolding (cmd/, internal/, go.mod)
- [ ] WorkItem, Phase, Status data models
- [ ] Store interface and MongoDB backend
- [ ] In-memory DAG engine (build, validate, readiness, critical path)
- [ ] Core CLI commands (init, next, start, done, block, unblock, reset, status)

### Up Next

- [ ] DAG evolution commands (sync, validate, add, remove)
- [ ] Markdown status report generation (`report`)
- [ ] Mermaid DAG visualization (`viz --format md`)
- [ ] PNG rendering -- local via chromedp (`viz --format png`)
- [ ] PNG rendering -- cloud via mermaid.ink (`viz --format png --renderer cloud`)
- [ ] `--agent-info` flag for agent self-onboarding
- [ ] `critical-path` command

---

## Phase 2: Distribution and Polish

### Cross-Platform Builds

- [ ] goreleaser configuration
- [ ] GitHub Actions CI/CD pipeline
- [ ] Multi-platform binaries (darwin/linux/windows, amd64/arm64)
- [ ] GitHub Releases automation

### Documentation

- [ ] Comprehensive README with examples
- [ ] `--agent-info` output refinement based on real agent usage
- [ ] Contributing guide

### Testing

- [ ] Unit tests for DAG engine (cycle detection, readiness, critical path)
- [ ] Unit tests for Store interface (MongoDB mock)
- [ ] Integration tests for CLI commands
- [ ] Test coverage target: 80%+

---

## Phase 3: Ecosystem Integration

### Cursor Integration

- [ ] Cursor rule template for consuming projects
- [ ] Cursor skill for Waypoint usage guidance
- [ ] Example seed files for common project structures

### Taskfile Integration

- [ ] Taskfile task templates for consuming projects
- [ ] `waypoint init` generates a starter Taskfile snippet

---

## Future Considerations

Items under consideration for future phases:

### GitHub Actions: Auto-Sync on Merge

A GitHub Action that runs `waypoint sync` automatically when a seed file is merged to `main`. This enables a GitOps workflow for work tracking:

1. Developer updates `seed/workitems.yaml` in a feature branch
2. PR is reviewed and merged to `main`
3. GitHub Action triggers `waypoint sync seed/workitems.yaml`
4. DAG in MongoDB is updated, new tasks become available to agents
5. Optionally runs `waypoint viz` and commits the updated diagram

```yaml
# .github/workflows/waypoint-sync.yml
name: Waypoint Sync
on:
  push:
    branches: [main]
    paths: ['seed/workitems.yaml']
jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: northharbor-dev/waypoint-action@v1
        with:
          command: sync seed/workitems.yaml
          mongo-uri: ${{ secrets.WAYPOINT_MONGO_URI }}
          project: ${{ github.event.repository.name }}
```

### Additional Storage Backends

- SQLite backend for zero-infrastructure local use
- PostgreSQL backend for teams preferring relational databases
- Neo4j backend for organizations wanting native graph query capabilities

### Enhanced Reporting

- HTML report output with embedded diagrams
- Velocity metrics (average task duration by role)
- Burndown data export (CSV/JSON) for external dashboards

### Webhook Notifications

- Notify on task state changes (Slack, Teams, Discord)
- Configurable per-project notification rules

### Multi-Tenancy

- Namespace isolation for org-wide deployments
- Role-based access control (who can reset/remove/force)
- Audit log of all state transitions

---

## How to Contribute

See the project README for setup instructions. For significant features, please create a design document first -- see [docs/design/](design/).

---

_Last updated: 2026-03-01_
