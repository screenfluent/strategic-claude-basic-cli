# Architecture Document
## Strategic Claude Basic CLI

### System Overview
Strategic Claude Basic CLI is a Go-based command-line tool that automates the setup and management of Strategic Claude Basic framework integration. The architecture follows standard Go CLI patterns using the Cobra framework for command structure and organization.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Interface                           │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────────┐ │
│  │  init   │  │  clean  │  │ status  │  │    completions     │ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                      Core Services                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Git Service   │  │ Installer Svc   │  │  Symlink Svc   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                    File System Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Directory Ops   │  │   File Ops      │  │  Symlink Ops   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
strategic-claude-basic-cli/
├── cmd/
│   └── strategic-claude-basic-cli/
│       └── main.go                    # Application entry point
├── internal/
│   ├── commands/                      # Command implementations
│   │   ├── root.go                    # Root command and global flags
│   │   ├── init.go                    # Init command
│   │   ├── clean.go                   # Clean command
│   │   ├── status.go                  # Status command
│   │   └── completions.go             # Completions command
│   ├── services/                      # Business logic services
│   │   ├── git.go                     # Git operations
│   │   ├── installer.go               # Installation logic
│   │   └── symlinks.go                # Symlink management
│   ├── models/                        # Data structures
│   │   ├── config.go                  # Configuration types
│   │   └── status.go                  # Status types
│   └── utils/                         # Utility functions
│       ├── fs.go                      # File system utilities
│       ├── paths.go                   # Path manipulation
│       └── validation.go              # Input validation
├── pkg/                               # Public interfaces (if needed)
├── docs/                             # Documentation
├── tests/                            # Test files
├── scripts/                          # Build and utility scripts
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── PRD.md
├── ARCHITECTURE.md                   # This file
└── ROADMAP.md
```

### Core Components

#### 1. CLI Layer (`internal/commands/`)
**Responsibility**: Handle user input, parse flags, coordinate command execution

- **Root Command**: Global configuration, help text, version info
- **Init Command**: Parse init-specific flags, validate input, orchestrate installation
- **Clean Command**: Handle cleanup flags, confirm destructive operations
- **Status Command**: Display current state information
- **Completions Command**: Generate shell completion scripts

#### 2. Service Layer (`internal/services/`)
**Responsibility**: Business logic and core operations

**Git Service** (`git.go`):
```go
type GitService struct {
    repoURL     string
    fixedCommit string
    tempDir     string
}

func (g *GitService) CloneAtCommit(destination string) error
func (g *GitService) GetTempDir() (string, error)
func (g *GitService) Cleanup() error
func (g *GitService) ValidateGitInstalled() error
```

**Installer Service** (`installer.go`):
```go
type InstallerService struct {
    gitService *GitService
    symlinkService *SymlinkService
}

func (i *InstallerService) AnalyzeInstallation(config InstallConfig) InstallationPlan
func (i *InstallerService) Install(config InstallConfig) error
func (i *InstallerService) InstallCore(config InstallConfig) error
func (i *InstallerService) Validate(targetDir string) error
func (i *InstallerService) GetStatus(targetDir string) StatusInfo
```

**Symlink Service** (`symlinks.go`):
```go
type SymlinkService struct{}

func (s *SymlinkService) CreateSymlinks(sourceDir, targetDir string) error
func (s *SymlinkService) RemoveSymlinks(targetDir string) error
func (s *SymlinkService) ValidateSymlinks(targetDir string) ([]SymlinkStatus, error)
```

**User Interaction Service** (`interaction.go`):
```go
type InteractionService struct{}

func (u *InteractionService) ConfirmInstallation(plan InstallationPlan) (bool, error)
func (u *InteractionService) FormatInstallationPrompt(plan InstallationPlan) string
```

#### 3. Model Layer (`internal/models/`)
**Responsibility**: Data structures and configuration

```go
type InstallConfig struct {
    TargetDir     string
    Force         bool
    ForceCore     bool
    SkipConfirm   bool
    NoBackup      bool
    Verbose       bool
}

type InstallationPlan struct {
    TargetDir        string
    InstallationType string  // "New Installation", "Update Core Only", "Full Overwrite"
    ExistingFiles    []string
    WillReplace      []string
    WillPreserve     []string
}

type StatusInfo struct {
    IsInstalled        bool
    StrategicClaudeDir bool
    ClaudeDir          bool
    Symlinks          []SymlinkStatus
    Issues            []string
}

type SymlinkStatus struct {
    Name   string
    Path   string
    Valid  bool
    Target string
}
```

#### 4. Utilities (`internal/utils/`)
**Responsibility**: Common utilities and helpers

- **File System Utils**: Directory creation, file copying, selective copying, existence checks
- **Path Utils**: Path resolution, relative path calculation
- **Validation Utils**: Input validation, permission checks
- **Selective Copy Utils**: Core directory identification, selective replacement logic
- **User Interaction Utils**: Confirmation prompts, input validation, terminal formatting

### Data Flow

#### Init Command Flow
1. **CLI Layer**: Parse flags, validate target directory
2. **Installer Service**: Analyze installation requirements
3. **CLI Layer**: Display confirmation prompt (unless `--yes`):
   - Show target directory path
   - Indicate installation type (New/Update Core/Full Overwrite)
   - Wait for user confirmation
4. **Git Service**: Clone repository to temp location and checkout fixed commit
5. **File System**: Copy `.strategic-claude-basic` to target:
   - **Full Installation** (`--force` or new): Copy entire directory
   - **Core Update** (`--force-core`): Selectively replace core/, guides/, templates/ (preserve archives/, issues/, plan/, research/, summary/)
6. **File System**: Create `.claude` directory structure (agents/, commands/, hooks/ subdirectories)
7. **Symlink Service**: Create symlinks:
   - `.claude/agents/strategic` → `../.strategic-claude-basic/core/agents`
   - `.claude/commands/strategic` → `../.strategic-claude-basic/core/commands`
   - `.claude/hooks/strategic` → `../.strategic-claude-basic/core/hooks`
8. **Git Service**: Cleanup temp directory
9. **Installer Service**: Validate installation

#### Clean Command Flow
1. **CLI Layer**: Parse flags, confirm destructive operation
2. **Symlink Service**: Remove strategic-claude symlinks
3. **File System**: Remove `.strategic-claude-basic` directory
4. **File System**: Clean up empty `.claude` directories

#### Status Command Flow
1. **CLI Layer**: Parse target directory
2. **Installer Service**: Check installation status
3. **Symlink Service**: Validate symlink integrity
4. **CLI Layer**: Format and display status information

### Directory Structure Strategy

#### Source Structure (strategic-claude-basic repository)
```
strategic-claude-basic/
└── .strategic-claude-basic/
    ├── archives/           # User workspace (preserved)
    ├── core/               # Framework core (updated)
    │   ├── agents/
    │   ├── commands/
    │   └── hooks/
    ├── guides/             # Framework docs (updated)
    ├── issues/             # User workspace (preserved)
    ├── plan/               # User workspace (preserved)
    ├── research/           # User workspace (preserved)
    ├── summary/            # User workspace (preserved)
    └── templates/          # Framework templates (updated)
```

#### Target Structure (after installation)
```
target-project/
├── .strategic-claude-basic/          # Copied from source
│   ├── archives/         # Preserved during --force-core
│   ├── core/             # Replaced during --force-core
│   │   ├── agents/
│   │   ├── commands/
│   │   └── hooks/
│   ├── guides/           # Replaced during --force-core
│   ├── issues/           # Preserved during --force-core
│   ├── plan/             # Preserved during --force-core
│   ├── research/         # Preserved during --force-core
│   ├── summary/          # Preserved during --force-core
│   └── templates/        # Replaced during --force-core
└── .claude/
    ├── agents/
    │   └── strategic -> ../../.strategic-claude-basic/core/agents
    ├── commands/
    │   └── strategic -> ../../.strategic-claude-basic/core/commands
    └── hooks/
        └── strategic -> ../../.strategic-claude-basic/core/hooks
```

#### Selective Update Strategy
**Core Directories** (replaced with `--force-core`):
- `core/` - Framework agents, commands, hooks
- `guides/` - Documentation and usage guides
- `templates/` - Framework templates and examples

**User Directories** (preserved with `--force-core`):
- `archives/` - User's archived work
- `issues/` - Project-specific issues tracking
- `plan/` - User's planning documents
- `research/` - User's research notes
- `summary/` - User's summaries and reports

### Technology Choices

#### Core Dependencies
- **Cobra**: CLI framework for command structure and flag parsing
- **Standard Library**: Maximize use of Go standard library for file operations
- **No external Git dependency**: Use `git` command via exec (ensures compatibility)

#### Development Dependencies
- **Go 1.21+**: Modern Go features and performance
- **Testing**: Standard Go testing framework
- **Linting**: golangci-lint for code quality
- **Build**: Makefile for build automation

### Error Handling Strategy

#### Error Categories
1. **User Input Errors**: Invalid flags, missing directories
2. **System Errors**: Permission issues, disk space
3. **Network Errors**: Git clone failures, connectivity issues
4. **State Errors**: Corrupted installations, broken symlinks

#### Error Handling Approach
- Return structured errors with context
- Provide actionable error messages
- Use appropriate exit codes
- Graceful cleanup on failures

```go
type CliError struct {
    Code    int
    Message string
    Cause   error
}

func (e *CliError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}
```

### Configuration Management

#### Runtime Configuration
- Command-line flags take precedence
- Environment variables for repository URL override
- Sensible defaults for all options

```go
const (
    // Repository information
    DefaultRepoURL = "git@github.com:Fomo-Driven-Development/strategic-claude-base.git"
    FixedCommit    = "db7a4c0"  // Pinned commit hash, updated with releases
    Branch         = "main"

    // Directory configuration
    DefaultTargetDir           = "."
    TempDirPrefix             = "strategic-claude-basic-"
    StrategicClaudeDir        = ".strategic-claude-basic"
    ClaudeDir                 = ".claude"
)
```

### Security Considerations
- Validate all file paths to prevent directory traversal
- Use relative symlinks for portability
- Proper cleanup of temporary directories
- No elevation of privileges required

### Performance Characteristics
- Git clone is the primary bottleneck (network dependent)
- Fixed commit checkout ensures consistent performance
- File operations are local and fast
- Symlink creation is near-instantaneous
- Status checks are very fast (just file system checks)

### Version Management Strategy
- **Fixed Commit Hash**: All installations use the same commit until CLI is updated
- **Manual Updates**: Framework version changes require new CLI releases
- **Consistency**: Prevents unexpected changes from repository updates
- **Stability**: Users get predictable, tested combinations of CLI and framework

### Testing Strategy
- Unit tests for each service component
- Integration tests for complete command flows
- Mock file system for testing edge cases
- Test coverage for error paths
