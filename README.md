# Espace-Tech Cloud CLI

Command-line tool for deploying and managing applications on [Espace-Tech Cloud](https://cloud.espace-tech.com).

## Installation

```bash
curl -fsSL https://api.espace-tech.com/install.sh | sh
```

Or download manually from [Releases](https://github.com/bz-reda/ETC-CLI/releases).

## Quick Start

```bash
espacetech register          # Create an account
espacetech login              # Authenticate
espacetech init               # Initialize a project
espacetech deploy --prod      # Deploy to production
```

## Commands

### Account

| Command | Description |
|---|---|
| `espacetech register` | Create a new Espace-Tech account |
| `espacetech login` | Authenticate with Espace-Tech Cloud (opens browser) |
| `espacetech login --email` | Authenticate with email/password |
| `espacetech logout` | Log out and clear saved credentials |
| `espacetech version` | Show CLI version |

### Projects

| Command | Description |
|---|---|
| `espacetech init` | Initialize a new project in the current directory |
| `espacetech deploy` | Deploy the current project (preview) |
| `espacetech deploy --prod` | Deploy to production |
| `espacetech status` | List your projects |
| `espacetech logs` | View application logs |
| `espacetech logs -n 500` | View last 500 lines |
| `espacetech rollback` | Rollback to a previous deployment |
| `espacetech delete` | Delete the current project and all its resources |

### Sites

| Command | Description |
|---|---|
| `espacetech site list` | List all sites in the current project |
| `espacetech site add [name]` | Add a new site to the current project |
| `espacetech site use <slug>` | Switch the active site in `.espacetech.json` |

### Domains

| Command | Description |
|---|---|
| `espacetech domain add [domain]` | Add a custom domain to the current project |
| `espacetech domain list` | List domains for the current project |
| `espacetech domain remove [domain]` | Remove a domain from the current project |

### Environment Variables

| Command | Description |
|---|---|
| `espacetech env set KEY=VALUE` | Set environment variables |
| `espacetech env set --file .env.production` | Set from file |
| `espacetech env list` | List environment variables |
| `espacetech env remove KEY` | Remove an environment variable |

### Databases

| Command | Description |
|---|---|
| `espacetech db create [name]` | Create a managed database |
| `espacetech db create [name] --type redis` | Create with specific type (postgres, redis, mongodb) |
| `espacetech db list` | List your databases |
| `espacetech db info [name]` | Show database details |
| `espacetech db credentials [name]` | Show connection credentials |
| `espacetech db link [name] --project [slug]` | Link database to a project (injects env vars) |
| `espacetech db unlink [name]` | Unlink database from its project |
| `espacetech db expose [name]` | Enable external access |
| `espacetech db unexpose [name]` | Disable external access |
| `espacetech db stop [name]` | Stop database (preserves data) |
| `espacetech db start [name]` | Start a stopped database |
| `espacetech db rotate [name]` | Rotate database password |
| `espacetech db delete [name]` | Delete database and all its data |

### Storage

| Command | Description |
|---|---|
| `espacetech storage create [name]` | Create a storage bucket |
| `espacetech storage list` | List your storage buckets |
| `espacetech storage info [name]` | Show bucket details |
| `espacetech storage credentials [name]` | Show S3 access credentials |
| `espacetech storage link [name] --project [slug]` | Link bucket to a project (injects S3 env vars) |
| `espacetech storage unlink [name]` | Unlink bucket from its project |
| `espacetech storage expose [name]` | Make bucket publicly accessible |
| `espacetech storage unexpose [name]` | Disable public access |
| `espacetech storage rotate [name]` | Rotate S3 access credentials |
| `espacetech storage delete [name]` | Delete bucket and all its data |

### Auth Apps

| Command | Description |
|---|---|
| `espacetech auth create [name]` | Create a managed auth service |
| `espacetech auth create [name] --app-id my-app` | Create with custom app ID |
| `espacetech auth list` | List your auth apps |
| `espacetech auth info [name]` | Show auth app details and endpoints |
| `espacetech auth config [name]` | Configure OAuth providers and settings |
| `espacetech auth users [name]` | List users for an auth app |
| `espacetech auth stats [name]` | Show auth app statistics |
| `espacetech auth rotate-keys [name]` | Rotate JWT signing keys |
| `espacetech auth delete [name]` | Delete auth app and all its users |

## Project Configuration

Running `espacetech init` creates a `.espacetech.json` file:

```json
{
  "project_id": "uuid",
  "name": "my-app",
  "slug": "my-app",
  "framework": "nextjs",
  "site_id": "uuid",
  "site_name": "main",
  "site_slug": "main"
}
```

## Building from Source

```bash
git clone https://github.com/bz-reda/ETC-CLI.git
cd ETC-CLI
make build
./espacetech version
```

## Documentation

Full documentation at [docs.espace-tech.com/cli](https://docs.espace-tech.com/cli).

## License

MIT
