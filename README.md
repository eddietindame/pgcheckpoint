# pgcheckpoint

A CLI tool for creating and restoring PostgreSQL database checkpoints using `pg_dump` and `psql`.

## Prerequisites

- Go 1.25+
- PostgreSQL client tools (`pg_dump` and `psql`) available in your `PATH`

## Installation

```sh
go install github.com/eddietindame/pgcheckpoint@latest
```

Or build from source:

```sh
make build
./bin/pgcheckpoint
```

## Usage

```sh
# Create a checkpoint (default command)
pgcheckpoint
pgcheckpoint create

# List checkpoints
pgcheckpoint list

# Restore database to latest checkpoint
pgcheckpoint restore

# Restore a specific checkpoint
pgcheckpoint restore checkpoint_2.sql

# Remove all but the latest checkpoint
pgcheckpoint prune
```

### Commands

| Command   | Description                                              |
| --------- | -------------------------------------------------------- |
| `create`  | Create a new checkpoint using `pg_dump`                  |
| `list`    | List all checkpoints for the active profile              |
| `restore` | Restore the database to a checkpoint (latest by default) |
| `prune`   | Remove all but the latest checkpoint                     |

Running `pgcheckpoint` without a subcommand defaults to `create`.

### Flags

```
-p, --port int              PostgreSQL port (default 5432)
    --db-user string        Database user (default "user")
    --db-password string    Database password (default "password")
    --db-host string        Database host (default "localhost")
    --db-name string        Database name (default "db")
    --db-sslmode string     SSL mode (default "disable")
    --checkpoint-dir string Checkpoint storage directory (default "~/.pgcheckpoint/checkpoints")
    --naming-mode string    Checkpoint naming mode: sequential, timestamp, compact, or unix (default "sequential")
-c, --config string         Global config file path
-j, --project-config string Project config file path
    --profile string        Config profile to use (default "default")
```

## Configuration

pgcheckpoint uses [Viper](https://github.com/spf13/viper) for configuration. Settings are resolved in the following order (highest priority first):

1. Command-line flags
2. Profile overrides
3. Project config (`.pgcheckpoint.yaml` in the current directory)
4. Global config
5. Environment variables
6. Default values

### Config file locations

**Global config** (searched in order):

- `~/.pgcheckpoint.yaml`
- `~/.pgcheckpoint/.pgcheckpoint.yaml`
- `~/.config/.pgcheckpoint.yaml`
- `~/.config/pgcheckpoint/.pgcheckpoint.yaml`

**Project config:**

- `./.pgcheckpoint.yaml`

### Example config

```yaml
db_user: dev
db_password: secret
db_host: localhost
db_port: 5432
db_name: myapp_dev
db_sslmode: disable
checkpoint_dir: /path/to/checkpoints
naming_mode: sequential # or "timestamp", "compact", "unix"

staging:
  db_user: staging_user
  db_password: staging_secret
  db_host: staging.example.com
  db_name: myapp_staging
  db_sslmode: require
```

Use the `--profile` flag to select a profile:

```sh
pgcheckpoint create --profile staging
```

### Checkpoint storage

Checkpoints are stored in your home directory by default:

```
~/.pgcheckpoint/checkpoints/{profile}/
```

For example: `~/.pgcheckpoint/checkpoints/default/checkpoint_1.sql`

The filename format depends on the active naming mode (see [Naming modes](#naming-modes)).

This can be overridden with the `--checkpoint-dir` flag or `checkpoint_dir` config key.

### Naming modes

The `--naming-mode` flag (or `naming_mode` config key) controls how checkpoint files are named:

| Mode | Example filename | Description |
| ------------ | ------------------------------------------ | ---------------------------------------- |
| `sequential` | `checkpoint_1.sql` | Numbered sequentially (default) |
| `timestamp` | `checkpoint_2026-03-02_15-30-45.sql` | Human-readable date and time |
| `compact` | `checkpoint_20260302T153045.sql` | RFC 3339 compact (no separators) |
| `unix` | `checkpoint_1740934245.sql` | Unix epoch seconds |

```sh
# Create checkpoints with different naming modes
pgcheckpoint create --naming-mode timestamp
pgcheckpoint create --naming-mode compact
pgcheckpoint create --naming-mode unix
```

You can also set this in your config file:

```yaml
naming_mode: timestamp
```

## Development

```sh
# Build
make build

# Run directly
make run ARGS="list --profile staging"

# Run tests
make test
```
