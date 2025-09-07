---
description: Update configuration for latest strategic-claude-basic-template version
allowed-tools: Bash, Read, Edit, WebFetch, TodoWrite, ExitPlanMode
---

# Template Update Automation

You are tasked with updating this CLI project's configuration to match the latest version of the strategic-claude-basic-template repository.

## Process Overview

**IMPORTANT: You MUST operate in plan mode first. Research thoroughly, then present your plan before making any changes.**

### Phase 1: Research (Read-only)
1. Clone the latest template to `/tmp/strategic-claude-basic-template-update`
2. Examine the current `FixedCommit` in `internal/config/constants.go`
3. Get the latest commit hash from the template repository
4. Compare directory structures between current config and template
5. Identify any new directories that need to be added to preserved/replaced lists
6. Check for any structural changes in the template

### Phase 2: Analysis & Planning
1. Create a todo list with all changes needed
2. Present detailed findings including:
   - Current vs latest commit hash
   - Any new directories discovered (e.g., validation/, product/)
   - Changes needed to constants.go
   - Documentation updates needed in CLAUDE.md
   - Any other configuration changes required

### Phase 3: Implementation (Only after plan approval)
1. Update `FixedCommit` in `internal/config/constants.go` to latest commit
2. Update `UserPreservedDirs` string if new user content directories found
3. Update `GetUserPreservedDirectories()` function if needed
4. Update CLAUDE.md documentation to reflect new directory structure
5. Verify all changes are consistent across the codebase

### Key Files to Check/Update:
- `internal/config/constants.go` - Commit hash and directory lists
- `CLAUDE.md` - Documentation of directory structure and framework behavior

### Template Repository Details:
- URL: https://github.com/Fomo-Driven-Development/strategic-claude-basic-template.git
- Branch: main
- Current location in code: `internal/config/constants.go`

**Remember**: Always research first, present a comprehensive plan, then execute only after approval.

Begin by researching the current state and latest template version.
