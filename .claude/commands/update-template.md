---
description: Update configuration for latest template version with multi-template support
allowed-tools: Bash, Read, Edit, WebFetch, TodoWrite, ExitPlanMode
---

# Multi-Template Update Automation

You are tasked with updating this CLI project's template configurations. The system now supports multiple templates (main, ccr, etc.) and you can update a specific template or choose interactively.

## Command Usage

```
/update-template [template-name]
```

**Examples:**
- `/update-template main` - Update main template only
- `/update-template ccr` - Update CCR template only
- `/update-template` - Interactive selection from available templates

## Process Overview

**IMPORTANT: You MUST operate in plan mode first. Research thoroughly, then present your plan before making any changes.**

### Phase 1: Template Selection
1. Parse the command input for template name
2. If template name provided:
   - Validate it exists in registry
   - Proceed with that template
3. If no template name provided:
   - Read `internal/templates/registry.go` to get available templates
   - Display list with current commits and descriptions
   - Ask user to select which template to update

### Phase 2: Research Selected Template (Read-only)
1. Clone the selected template's branch to `/tmp/strategic-claude-base-update`
2. Get the latest commit hash from the template's specific branch
3. Compare with current commit in `internal/templates/registry.go`
4. Check for any structural changes in the selected template
5. Identify any changes needed to directory structure constants

### Phase 3: Analysis & Planning
1. Create a todo list with all changes needed
2. Present detailed findings including:
   - Current vs latest commit hash for selected template
   - Any new directories discovered
   - Changes needed to registry.go
   - Impact on other templates (should be none)
   - Documentation updates needed

### Phase 4: Implementation (Only after plan approval)
1. Update the selected template's commit hash in `internal/templates/registry.go`
2. Keep other templates' commits unchanged
3. Update directory constants if new directories found
4. Update CLAUDE.md documentation if structural changes detected
5. Run tests to verify changes work correctly

### Key Files to Check/Update:
- `internal/templates/registry.go` - Template registry with commit hashes
- `internal/config/constants.go` - Directory lists (only if structural changes)
- `CLAUDE.md` - Documentation (only if structural changes)

### Template Registry Location:
- File: `internal/templates/registry.go`
- Contains: Template definitions with branches and commit hashes

### Available Templates:
Templates are defined in the registry with their respective branches:
- **main**: Strategic Claude Basic (branch: main)
- **ccr**: CCR Template (branch: ccr-template)

**Remember**:
1. Always research first, present a comprehensive plan, then execute only after approval
2. Only update the selected template's commit hash
3. Keep other templates stable at their current commits
4. Test changes after implementation

Begin by determining which template to update, then research the current state vs latest version.
