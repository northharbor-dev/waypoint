# Getting Started with Waypoint

This guide walks you through setting up Waypoint from scratch: provisioning a MongoDB Atlas cluster, installing the CLI, seeding a project, and running your first workflow.

## Prerequisites

- **Go 1.22+** (for building from source) or download a pre-built binary
- **MongoDB Atlas account** -- the free M0 tier is sufficient
- **Git**

## Step 1: MongoDB Atlas Setup

Waypoint uses MongoDB for persistence. Atlas provides a free hosted cluster that works out of the box.

### Create a Free Cluster

1. Go to [cloud.mongodb.com](https://cloud.mongodb.com) and sign up or log in.
2. Create a new project (e.g., "waypoint").
3. Click **Build a Database** and select the **M0 (Free)** tier.
4. Choose any cloud provider and region.
5. Name your cluster (e.g., "waypoint-cluster").
6. Click **Create Deployment**.

Provisioning takes 1-3 minutes. You can configure access while it completes.

### Configure Network Access

1. In the left sidebar, go to **Network Access**.
2. Click **Add IP Address**.
3. For development, select **Allow Access from Anywhere** (`0.0.0.0/0`).
4. For production, restrict to specific IPs.
5. Click **Confirm**.

### Create a Database User

1. In the left sidebar, go to **Database Access**.
2. Click **Add New Database User**.
3. Choose **Password** authentication.
4. Enter a username (e.g., `waypoint-user`).
5. Generate or create a strong password. Save it -- you'll need it for the connection string.
6. Under **Database User Privileges**, select **Read and write to any database**.
7. Click **Add User**.

### Get Your Connection String

1. In the left sidebar, go to **Database**.
2. Click **Connect** on your cluster.
3. Select **Drivers**.
4. Copy the connection string. It looks like:
   ```
   mongodb+srv://waypoint-user:<password>@waypoint-cluster.abc123.mongodb.net/?retryWrites=true&w=majority
   ```
5. Replace `<password>` with your actual password.

## Step 2: Install Waypoint

### From Source

```bash
go install github.com/northharbor-dev/waypoint/cmd/waypoint@latest
```

### From Releases

Download the binary for your platform from [Releases](https://github.com/northharbor-dev/waypoint/releases) and place it on your `PATH`.

### Verify

```bash
waypoint --help
```

Expected output:

```
Graph-based work tracking for AI-assisted development

Usage:
  waypoint [flags]
  waypoint [command]

Available Commands:
  add           Add a work item
  block         Mark a task as blocked
  completion    Generate the autocompletion script for the specified shell
  critical-path Show the longest dependency chain
  done          Mark a task as complete
  help          Help about any command
  init          Seed a new project from YAML
  next          Show tasks ready to start
  remove        Remove a work item
  report        Generate a status report
  reset         Release an orphaned task
  start         Claim a task
  status        Show all tasks with current state
  sync          Reconcile updated YAML with live state
  unblock       Clear a blocker
  validate      Check DAG integrity
  viz           Generate a DAG visualization
```

## Step 3: Configure Environment

Create a `.env` file in your project root (see [`.env.example`](../.env.example) for reference):

```
WAYPOINT_MONGO_URI=mongodb+srv://waypoint-user:yourpassword@waypoint-cluster.abc123.mongodb.net/?retryWrites=true&w=majority
WAYPOINT_MONGO_DATABASE=waypoint
WAYPOINT_PROJECT=my-project
```

Or export the variables directly:

```bash
export WAYPOINT_MONGO_URI="mongodb+srv://waypoint-user:yourpassword@waypoint-cluster.abc123.mongodb.net/?retryWrites=true&w=majority"
export WAYPOINT_MONGO_DATABASE="waypoint"
export WAYPOINT_PROJECT="my-project"
```

Waypoint loads `.env` automatically if present. `WAYPOINT_MONGO_URI` and `WAYPOINT_PROJECT` are required; `WAYPOINT_MONGO_DATABASE` defaults to `waypoint`.

## Step 4: Create Your Work Breakdown

Create a `seed/workitems.yaml` file defining your project's phases and tasks:

```yaml
project: my-project
phases:
  - number: 1
    name: Foundation
    target: "Q2 2026"
  - number: 2
    name: Integration
    target: "Q3 2026"

work_items:
  # Phase 1 -- Foundation
  - id: WI-001
    title: "Define data model and schemas"
    phase: 1
    owner: human
    role: lead
    dependencies: []
  - id: WI-002
    title: "Implement database layer"
    phase: 1
    owner: agent
    role: data_model
    dependencies: [WI-001]
  - id: WI-003
    title: "Build REST API endpoints"
    phase: 1
    owner: agent
    role: api_dev
    dependencies: [WI-002]
  - id: WI-004
    title: "Write unit tests for API"
    phase: 1
    owner: agent
    role: unit_test
    dependencies: [WI-003]

  # Phase 2 -- Integration
  - id: WI-005
    title: "Build dashboard UI"
    phase: 2
    owner: agent
    role: ui_dev
    dependencies: [WI-003]
  - id: WI-006
    title: "End-to-end integration tests"
    phase: 2
    owner: agent
    role: integration
    dependencies: [WI-004, WI-005]
```

Key points:
- **`owner`** is `human` or `agent` -- indicates who performs the work.
- **`role`** determines which agent (or person) can claim the task. Valid roles: `lead`, `api_dev`, `ui_dev`, `backend_dev`, `security`, `unit_test`, `integration`, `data_model`, `scenario`.
- **`dependencies`** lists work item IDs that must be `done` before this item becomes available.

## Step 5: Seed Your Project

```bash
waypoint init seed/workitems.yaml
```

```
Seeded project "my-project": 6 work items, 2 phases
```

Waypoint creates the `work_items` and `phases` collections in your MongoDB database. Run `validate` to confirm the DAG is well-formed:

```bash
waypoint validate
```

```
DAG is valid: 6 items, 0 issues
```

## Step 6: Start Working

The core workflow is three commands: `next`, `start`, `done`.

**Find available work for a role:**

```bash
waypoint next --role data_model
```

```
Ready tasks for role "data_model":
  WI-002  Implement database layer  (phase 1)
```

**Claim the task:**

```bash
waypoint start WI-002
```

```
Started WI-002: Implement database layer
```

The claim is atomic -- if two agents race for the same task, only one succeeds. The other gets an "already claimed" error and should run `next` again.

**Complete the task:**

```bash
waypoint done WI-002
```

```
Completed WI-002: Implement database layer (duration: 2h 15m)
```

Downstream tasks that depended on WI-002 now become available:

```bash
waypoint next --role api_dev
```

```
Ready tasks for role "api_dev":
  WI-003  Build REST API endpoints  (phase 1)
```

**If something is blocked:**

```bash
waypoint block WI-003 -r "Waiting on API contract review"
```

```
Blocked WI-003: Build REST API endpoints
  Reason: Waiting on API contract review
```

```bash
waypoint unblock WI-003
```

```
Unblocked WI-003: Build REST API endpoints (status: not_started)
```

## Step 7: Check Progress

**View all tasks and their current state:**

```bash
waypoint status
```

```
my-project status:

Phase 1: Foundation (target: Q2 2026)
  WI-001  ✓ done          Define data model and schemas         lead
  WI-002  ✓ done          Implement database layer              data_model
  WI-003  ● in_progress   Build REST API endpoints              api_dev
  WI-004  ○ not_started   Write unit tests for API              unit_test

Phase 2: Integration (target: Q3 2026)
  WI-005  ○ not_started   Build dashboard UI                    ui_dev
  WI-006  ○ not_started   End-to-end integration tests          integration

Progress: 2/6 done (33%)
```

**Filter by phase:**

```bash
waypoint status --phase 1
```

**Export as JSON for programmatic use:**

```bash
waypoint status --output json
```

**Generate a markdown status report:**

```bash
waypoint report --output status-report.md
```

**Visualize the dependency graph:**

```bash
waypoint viz --format md --output dag.md
```

**Find the longest dependency chain:**

```bash
waypoint critical-path
```

```
Critical path (4 items):
  WI-001 → WI-002 → WI-003 → WI-004
```

## Next Steps

- **AI agents:** Run `waypoint --agent-info` for a self-contained onboarding guide that agents can consume directly -- no external docs needed.
- **Cursor rules:** Set up a `.cursor/rules/` file that tells agents to run `waypoint next --role <role>` to find their tasks.
- **Taskfile:** Add a `Taskfile.yml` with shortcuts like `task work:next` for human convenience.
- **Deeper understanding:** Read [architecture.md](architecture.md) for the data model, concurrency model, and DAG validation details.
- **Evolving the plan:** Edit your `seed/workitems.yaml` and run `waypoint sync seed/workitems.yaml` to reconcile changes with live state without losing progress.
