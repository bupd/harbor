# 8gcr-ee Patch Management

This directory contains patches for 8gcr Enterprise Edition features on top of OSS Harbor.

See [ADR-0001](../decision-records/0001-managing-8gcr-modifications-on-harbor.md) for background and design decisions.

## Prerequisites

Install stgit:

```bash
brew install stgit
```

## Quick Reference

| Task | Command |
|------|---------|
| Initialize | `stg init` |
| Import patches | `stg import -S 8gcr-ee/patches/series` |
| List stack | `stg series` |
| Go to patch | `stg goto <patch-name>` |
| Create new patch | `stg new <name> -m "message"` |
| Update current patch | `stg refresh` |
| Apply remaining patches | `stg push -a` |
| Export to files | `stg export -d 8gcr-ee/patches/` |
| Rebuild patch from commit | `git cherry-pick --no-commit <hash>` then `stg new` + `stg refresh` |
| Rebase onto source branch | `stg rebase <patch-source-branch>` |

## Terminology

- **Patch source branch** (e.g., `main`): stores patch files as the source of truth.
- **Wip branch** (e.g., `wip/my-feature`): has patches applied for development. Temporary and local — never push or commit applied results.

## Workflows

### Applying Patches for Development

```bash
git checkout main -b wip/applied
stg init
stg import -S 8gcr-ee/patches/series
# Work on wip/applied branch - DO NOT push or commit result
```

### Fix a Failing Patch (Manual Apply)

```bash
# If stg import fails on a patch:
# 1. Resolve the conflict
git add <resolved-files>
stg refresh
# 2. Continue importing remaining patches
stg import -S 8gcr-ee/patches/series  # continues from where it left off
```

### Making Changes to an Existing Patch

```bash
stg goto 0002-ldap-admin-group-filter  # Jump to that patch
# Make your changes...
git add -A
stg refresh                            # Updates the patch
stg push -a                            # Re-apply remaining patches
stg export -d 8gcr-ee/patches/         # Export updated patches
# Then commit the updated patch files to main
```

### Creating a New Feature Patch

```bash
# Ensure all patches are applied
stg push -a

# Create a new patch
stg new 0013-my-feature -m "feat(scope): add my feature"

# Make your changes
# ...

# Update the patch
stg refresh

# Export patches (series file is updated automatically)
stg export -d 8gcr-ee/patches/
```

### Export Patches After Changes

```bash
stg export -d 8gcr-ee/patches/
git checkout main
git add 8gcr-ee/patches/
git commit -m "fix(patches): <description>"
```

### Rebuild a Patch via Cherry-Pick

Use this workflow when:
- A patch file is malformed (wrong format, corrupted, or fails to import)
- You resolved merge conflicts in the wip branch and need to update the patch files in the patch source branch

```bash
# 1. Ensure you're on a wip branch with StGit initialized and prior patches applied
stg series  # verify current stack state

# 2. Cherry-pick the commit without committing (stages changes only)
git cherry-pick --no-commit <commit-hash>

# 3. Resolve any conflicts if they occur
git status                    # check for conflicts
# ... resolve conflicts ...
git add <resolved-files>

# 4. Create a new StGit patch from the staged changes
stg new <patch-name> -m "feat(scope): description"
stg refresh

# 5. Continue with remaining patches if any
stg import <next-patch>       # or stg push if already in stack

# 6. Export all patches to the patch source branch
stg export -d 8gcr-ee/patches/

# 7. Commit updated patches in the patch source branch
cd <patch-source-worktree>
git add 8gcr-ee/patches/
git commit -m "fix(patches): rebuild <patch-name> with conflict resolution"

# 8. Rebase wip branch onto updated patch source branch
cd <wip-worktree>
stg rebase <patch-source-branch>
```

### Rebasing Patches to New OSS Version

```bash
# Pop all patches
stg pop -a

# Rebase to new version
stg rebase v2.14.0

# Re-apply patches, fixing conflicts
stg push -a
# If conflicts occur:
#   1. Resolve conflicts
#   2. git add resolved files
#   3. stg refresh
#   4. stg push (continue)

# Export updated patches
stg export -d 8gcr-ee/patches/
```

## File Structure

```
8gcr-ee/
└── patches/
    ├── series                    # Patch ordering (stgit format)
    ├── README.md                 # This file
    ├── 0001-branding
    ├── 0002-ldap-admin-group-filter
    └── ...
```

## Series File Format

The `series` file lists patches in application order:
- One patch filename per line
- Lines starting with `#` are comments
- Blank lines are ignored
- Patches are applied in order from top to bottom

## Patch Naming Convention

```
NNNN-short-description
```

- `NNNN`: 4-digit sequence number (0001, 0002, etc.)
- `short-description`: Kebab-case description of the feature
- No `.patch` file extension (StGit export matches names directly)

## Tips

- Keep patches focused on a single feature
- Use `stg series -d` to see patch descriptions
- Use `stg show` to view the current patch diff
- Use `stg log` to see patch history
- Use `stg refresh --force` if the index is dirty from a `cherry-pick --no-commit`
