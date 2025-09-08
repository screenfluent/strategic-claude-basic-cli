# Product Requirements Document (PRD)
## Strategic Claude Basic CLI

### Overview
Strategic Claude Basic CLI is a command-line tool that simplifies the integration of the Strategic Claude Basic framework into any project. It automates the process of cloning the strategic-claude-base repository, copying its contents, and setting up the necessary directory structure and symlinks for Claude Code integration.

### Goals
- Provide a simple, reliable way to add Strategic Claude Basic to any project
- Automate the complex setup process of directory structures and symlinks
- Enable easy management and cleanup of Strategic Claude Basic installations
- Support standard CLI patterns and shell completions

### Target Users
- **Software Developers**: Using Claude Code for development and wanting to add Strategic Claude Basic capabilities
- **Project Maintainers**: Setting up consistent Claude Code configurations across repositories
- **DevOps Engineers**: Automating development environment setup

### Core Use Cases

#### 1. Initialize Strategic Claude Basic in a project
**User Story**: As a developer, I want to add Strategic Claude Basic to my project with a single command so I can start using its features immediately.

**Acceptance Criteria**:
- Clone strategic-claude-base repository to temporary location
- Copy `.strategic-claude-basic` directory to target project
- Create `.claude` directory if it doesn't exist
- Set up symlinks from `.claude` subdirectories (agents/, commands/, hooks/) to strategic-claude-basic core components
- Clean up temporary files
- Provide clear feedback on success/failure

#### 1.1. Update core Strategic Claude Basic content
**User Story**: As a developer with an existing Strategic Claude Basic installation, I want to update the core framework content without losing my project-specific customizations.

**Acceptance Criteria**:
- Selectively replace only framework directories (core, guides, templates)
- Preserve user directories (archives, issues, plan, research, summary)
- Maintain existing symlinks and configuration
- Provide clear feedback on what was updated vs. preserved

#### 2. Check installation status
**User Story**: As a developer, I want to verify that Strategic Claude Basic is properly installed and configured in my project.

**Acceptance Criteria**:
- Check if `.strategic-claude-basic` directory exists
- Verify symlinks are valid and point to correct locations
- Display current installation state clearly
- Identify any configuration issues

#### 3. Remove Strategic Claude Basic from project
**User Story**: As a project maintainer, I want to cleanly remove Strategic Claude Basic from a project when it's no longer needed.

**Acceptance Criteria**:
- Remove `.strategic-claude-basic` directory
- Remove strategic-claude symlinks from `.claude` subdirectories
- Optionally clean up empty `.claude` directory
- Provide confirmation before destructive operations

#### 4. Shell completion support
**User Story**: As a CLI user, I want tab completion for commands and flags to improve my productivity.

**Acceptance Criteria**:
- Generate completion scripts for bash, zsh, fish, powershell
- Support completion of command names and flags
- Provide installation instructions for completions

### Command Specifications

#### `strategic-claude-basic-cli init`
**Purpose**: Initialize Strategic Claude Basic in a project

**Flags**:
- `--target, -t`: Target directory (default: current directory)
- `--force, -f`: Overwrite existing installation completely
- `--force-core`: Update only framework directories (core, guides, templates) while preserving user content (archives, issues, plan, research, summary)
- `--yes, -y`: Skip confirmation prompt (auto-confirm installation)
- `--no-backup`: Skip backup creation
- `--verbose, -v`: Verbose output

**Behavior**:
1. Validate target directory exists and is writable
2. Check for existing installation (unless `--force` or `--force-core`)
3. Display installation summary and prompt for confirmation (unless `--yes`):
   ```
   Target directory: /path/to/project
   Installation type: [New Installation | Update Core Only | Full Overwrite]

   This will install Strategic Claude Basic in the above directory.
   Are you sure you want to proceed? (y/N):
   ```
4. Create backup of existing files (unless `--no-backup`)
5. Clone strategic-claude-base repository
6. Copy `.strategic-claude-basic` to target:
   - If `--force-core`: Only replace core/, guides/, templates/ directories (preserve archives/, issues/, plan/, research/, summary/)
   - If `--force`: Replace entire directory
   - Otherwise: Full installation (fail if exists)
7. Create `.claude` directory structure (agents/, commands/, hooks/ subdirectories)
8. Set up symlinks:
   - `.claude/agents/strategic` → `../.strategic-claude-basic/core/agents`
   - `.claude/commands/strategic` → `../.strategic-claude-basic/core/commands`
   - `.claude/hooks/strategic` → `../.strategic-claude-basic/core/hooks`
9. Validate installation
10. Clean up temporary files

#### `strategic-claude-basic-cli clean`
**Purpose**: Remove Strategic Claude Basic from a project

**Flags**:
- `--target, -t`: Target directory (default: current directory)
- `--force, -f`: Skip confirmation prompts
- `--verbose, -v`: Verbose output

**Behavior**:
1. Validate target directory
2. Check for existing installation
3. Prompt for confirmation (unless `--force`)
4. Remove symlinks
5. Remove `.strategic-claude-basic` directory
6. Clean up empty `.claude` directories if appropriate

#### `strategic-claude-basic-cli status`
**Purpose**: Display installation status

**Flags**:
- `--target, -t`: Target directory (default: current directory)

**Behavior**:
1. Check if `.strategic-claude-basic` exists
2. Verify symlink integrity
3. Display installation status with clear indicators
4. Report any issues found

#### `strategic-claude-basic-cli completions`
**Purpose**: Generate shell completion scripts

**Arguments**:
- `shell`: Shell type (bash|zsh|fish|powershell)

**Behavior**:
1. Generate appropriate completion script for specified shell
2. Output to stdout for easy installation
3. Provide installation instructions

### Technical Requirements

#### System Requirements
- Git installed and available in PATH
- Go 1.21+ for building from source
- Write permissions to target directories

#### Version Management
- Strategic Claude Basic content is pinned to a specific commit hash
- All installations use the same framework version until CLI is updated
- Framework updates require new CLI releases for consistency
- No user-configurable version selection to maintain stability

#### Performance Requirements
- Init command should complete within 30 seconds on typical network
- Status check should be near-instantaneous
- Minimal resource usage

#### Error Handling
- Clear, actionable error messages
- Graceful handling of network issues
- Proper cleanup on interruption
- Non-zero exit codes for failures
- User-friendly confirmation prompts with clear defaults

#### User Experience
- Interactive confirmation prevents accidental installations
- Clear indication of installation type and target directory
- `--yes` flag enables automation while maintaining safety by default

### Success Metrics
- Users can successfully initialize Strategic Claude Basic in under 30 seconds
- Zero failed installations due to permission or symlink issues
- Clear documentation and help text for all commands
- Shell completions work across major shells

### Non-Goals
- Complex configuration management beyond basic installation
- Integration with other Claude frameworks
- GUI interface
- Automatic updates of strategic-claude-base content
