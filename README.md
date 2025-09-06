# Strategic Claude Basic CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/dgnsrekt/strategic-claude-basic-cli/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A command-line tool that simplifies the integration of the [Strategic Claude Basic framework](https://github.com/Fomo-Driven-Development/strategic-claude-basic-template) into your development projects.

## Overview

Strategic Claude Basic CLI automates the complex setup process of integrating Strategic Claude Basic framework with Claude Code. Instead of manually cloning repositories, copying directories, and creating symlinks, this tool handles everything with a single command.

### Key Features

- **One-command installation** - Set up Strategic Claude Basic in any project instantly
- **Smart updates** - Update framework core while preserving your custom content
- **Automatic symlinks** - Creates proper `.claude` integration for Claude Code
- **Status validation** - Verify installation health and detect issues
- **Clean removal** - Safely remove framework without affecting your work
- **Shell completions** - Tab completion for bash, zsh, fish, and PowerShell
- **Zero dependencies** - Uses only Git and Go standard library

### What it does

1. **Clones** Strategic Claude Basic template at a fixed, tested commit
2. **Installs** `.strategic-claude-basic/` directory with framework content
3. **Creates** `.claude/` directory structure for Claude Code integration
4. **Links** Claude directories to framework components via symlinks
5. **Preserves** your custom content during updates

## Installation

### Prerequisites

- **Git** - Must be installed and available in your PATH
- **Go 1.21+** - Required for building from source

### Install with Go

```bash
go install github.com/dgnsrekt/strategic-claude-basic-cli/cmd/strategic-claude-basic-cli@latest
```

### Build from source

```bash
git clone https://github.com/dgnsrekt/strategic-claude-basic-cli.git
cd strategic-claude-basic-cli
make build
# Binary will be available at ./bin/strategic-claude-basic-cli
```

### Install binary to PATH

```bash
make install  # Installs to $GOPATH/bin or $HOME/go/bin
```

## Quick Start

1. **Initialize** Strategic Claude Basic in your project:
   ```bash
   strategic-claude-basic-cli init
   ```

2. **Check status** of your installation:
   ```bash
   strategic-claude-basic-cli status
   ```

3. **Update** framework core files (preserves your custom content):
   ```bash
   strategic-claude-basic-cli init --force-core
   ```

## Usage

### Initialize Framework (`init`)

Install Strategic Claude Basic in a new or existing project:

```bash
# Install in current directory
strategic-claude-basic-cli init

# Install in specific directory
strategic-claude-basic-cli init ./my-project

# Preview what would be installed (dry run)
strategic-claude-basic-cli init --dry-run

# Install with auto-confirmation
strategic-claude-basic-cli init --yes
```

**Update existing installations:**

```bash
# Update only framework files, preserve user content
strategic-claude-basic-cli init --force-core

# Completely overwrite existing installation
strategic-claude-basic-cli init --force
```

### Check Status (`status`)

Verify your installation and diagnose issues:

```bash
# Check current directory
strategic-claude-basic-cli status

# Check specific directory
strategic-claude-basic-cli status --target ./my-project

# Verbose output with detailed diagnostics
strategic-claude-basic-cli status --verbose
```

### Clean Installation (`clean`)

Remove Strategic Claude Basic from your project:

```bash
# Remove with confirmation prompt
strategic-claude-basic-cli clean

# Force removal without confirmation
strategic-claude-basic-cli clean --force

# Clean specific directory
strategic-claude-basic-cli clean ./my-project
```

### Shell Completions (`completions`)

Set up tab completion for your shell:

```bash
# Generate completions for your shell
strategic-claude-basic-cli completions bash
strategic-claude-basic-cli completions zsh
strategic-claude-basic-cli completions fish
strategic-claude-basic-cli completions powershell

# Install bash completions (example)
strategic-claude-basic-cli completions bash > /usr/local/etc/bash_completion.d/strategic-claude-basic-cli
```

## Directory Structure

After installation, your project will have this structure:

```
your-project/
├── .strategic-claude-basic/          # Framework installation
│   ├── core/                        # Framework components (updated)
│   │   ├── agents/                  # Claude agents
│   │   ├── commands/                # Claude commands
│   │   └── hooks/                   # Claude hooks
│   ├── guides/                      # Framework documentation (updated)
│   ├── templates/                   # Framework templates (updated)
│   ├── archives/                    # Your archived work (preserved)
│   ├── issues/                      # Your issues tracking (preserved)
│   ├── plan/                        # Your planning docs (preserved)
│   ├── research/                    # Your research notes (preserved)
│   └── summary/                     # Your summaries (preserved)
└── .claude/                         # Claude Code integration
    ├── agents/
    │   └── strategic -> ../../.strategic-claude-basic/core/agents
    ├── commands/
    │   └── strategic -> ../../.strategic-claude-basic/core/commands
    └── hooks/
        └── strategic -> ../../.strategic-claude-basic/core/hooks
```

### Framework vs User Content

**Framework directories** (replaced during `--force-core` updates):
- `core/` - Strategic Claude agents, commands, and hooks
- `guides/` - Framework documentation and usage guides
- `templates/` - Framework templates and examples

**User directories** (preserved during `--force-core` updates):
- `archives/` - Your archived work and completed projects
- `issues/` - Project-specific issue tracking
- `plan/` - Your planning documents and strategies
- `research/` - Your research notes and findings
- `summary/` - Your summaries and reports

## Installation Types

### New Installation
For projects without Strategic Claude Basic:
```bash
strategic-claude-basic-cli init
```
- Installs complete `.strategic-claude-basic/` directory
- Creates `.claude/` symlinks
- Safe - fails if installation already exists

### Core Update (`--force-core`)
For updating framework while preserving your work:
```bash
strategic-claude-basic-cli init --force-core
```
- Updates `core/`, `guides/`, `templates/` directories
- Preserves `archives/`, `issues/`, `plan/`, `research/`, `summary/`
- Maintains your custom content and configurations

### Full Overwrite (`--force`)
For complete reinstallation:
```bash
strategic-claude-basic-cli init --force
```
- Replaces entire `.strategic-claude-basic/` directory
- **Warning**: This will overwrite all your custom user content
- Creates backup unless `--no-backup` is specified

## Commands Reference

| Command | Purpose | Key Flags |
|---------|---------|-----------|
| `init` | Install/update Strategic Claude Basic | `--force-core`, `--force`, `--yes`, `--dry-run` |
| `status` | Check installation health | `--verbose` |
| `clean` | Remove Strategic Claude Basic | `--force` |
| `completions` | Generate shell completions | Shell type argument |
| `version` | Show version information | - |

For detailed help on any command:
```bash
strategic-claude-basic-cli [command] --help
```

## Development

### Building

```bash
# Build binary
make build

# Build and run
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
make test-coverage
```

## Version Management

Strategic Claude Basic CLI pins the framework to specific, tested commits to ensure stability:
- All installations use the same framework version
- Framework updates require new CLI releases
- No unexpected changes from upstream repository updates
- Consistent, predictable behavior across installations

## Related Documentation

- **[Architecture](ARCHITECTURE.md)** - Detailed system design and technical architecture
- **[Product Requirements](PRD.md)** - Features, use cases, and specifications
- **[Development Roadmap](ROADMAP.md)** - Development phases and progress tracking
- **[Claude Integration](CLAUDE.md)** - Instructions for working with Claude Code

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: Check the docs in this repository
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/dgnsrekt/strategic-claude-basic-cli/issues)
- **Discussions**: Join conversations on [GitHub Discussions](https://github.com/dgnsrekt/strategic-claude-basic-cli/discussions)
