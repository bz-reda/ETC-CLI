# Espace-Tech Cloud CLI

Command-line tool for deploying and managing applications on [Espace-Tech Cloud](https://cloud.espace-tech.com).

## Installation

```bash
curl -fsSL https://cloud.espace-tech.com/install.sh | sh
```

Or download manually from [Releases](https://github.com/bz-reda/ETC-CLI/releases).

## Quick Start

```bash
# Login to your account
espacetech login

# Initialize a project
espacetech init

# Deploy
espacetech deploy
```

## Commands

| Command | Description |
|---------|-------------|
| `espacetech login` | Authenticate with Espace-Tech Cloud |
| `espacetech init` | Initialize a new project |
| `espacetech deploy` | Deploy your application |
| `espacetech logs` | View application logs |
| `espacetech domains add` | Add a custom domain |
| `espacetech env set` | Set environment variables |
| `espacetech db create` | Create a managed database |
| `espacetech rollback` | Rollback to a previous deployment |
| `espacetech delete` | Delete a project |
| `espacetech version` | Show CLI version |

For full documentation, visit [docs.cloud.espace-tech.com/cli](https://docs.cloud.espace-tech.com/cli).

## Building from Source

```bash
git clone https://github.com/bz-reda/ETC-CLI.git
cd ETC-CLI
make build
./espacetech version
```

## License

MIT © [Espace-Tech](https://espace-tech.com)
