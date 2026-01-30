# 8gcr-ee Patch Management

This directory contains patches for 8gcr Enterprise Edition features on top of OSS Harbor.

See [ADR-0001](../../docs/adr/0001-managing-8gcr-modifications-on-harbor.md) for background and design decisions.

## Prerequisites

Install stgit:

```bash
brew install stgit
```

## Quick Reference

| Task | Command |
|------|---------|
| Initialize | `stg init` |
| Import patches | `stg import -s 8gcr-ee/patches/series` |
| Apply to specific | `stg goto <patch>` |
| List stack | `stg series` |
| Create new patch | `stg new <name> -m "message"` |
| Update patch | `stg refresh` |
| Export to files | `stg export -d 8gcr-ee/patches/ -n` |

## Workflows

### Applying Patches to Fresh OSS Harbor

```bash
# Checkout OSS Harbor at desired version
git checkout v2.12.0

# Initialize stgit
stg init

# Import all patches
stg import -s 8gcr-ee/patches/series
```

### Making Changes to an Existing Patch

```bash
# Go to the patch you want to modify
stg goto 0003-hybrid-auth-multi-source

# Make your changes to the code
# ...

# Update the patch with your changes
stg refresh

# Re-apply remaining patches
stg push -a

# Export updated patches
stg export -d 8gcr-ee/patches/ -n
```

### Creating a New Feature Patch

```bash
# Ensure all patches are applied
stg push -a

# Create a new patch
stg new 0013-my-feature -m "Add my feature"

# Make your changes
# ...

# Update the patch
stg refresh

# Export patches
stg export -d 8gcr-ee/patches/ -n

# Add to series file
echo "0013-my-feature.patch" >> 8gcr-ee/patches/series
```

### Rebasing Patches to New OSS Version

```bash
# Pop all patches
stg pop -a

# Rebase to new version
git rebase v2.13.0

# Re-apply patches, fixing conflicts
stg push -a
# If conflicts occur:
#   1. Resolve conflicts
#   2. git add resolved files
#   3. stg refresh
#   4. stg push (continue)

# Export updated patches
stg export -d 8gcr-ee/patches/ -n
```

## File Structure

```
8gcr-ee/
└── patches/
    ├── series                    # Patch ordering (stgit format)
    ├── README.md                 # This file
    ├── 0001-cr-init-build-setup.patch
    ├── 0002-ldap-admin-group-filter.patch
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
NNNN-short-description.patch
```

- `NNNN`: 4-digit sequence number (0001, 0002, etc.)
- `short-description`: Kebab-case description of the feature

## Tips

- Keep patches focused on a single feature
- Schema migrations (0012) should always be last in series
- Use `stg series -d` to see patch descriptions
- Use `stg show` to view the current patch diff
- Use `stg log` to see patch history
