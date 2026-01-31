---
name: stgit
description: Manage 8gcr-ee patch stack on OSS Harbor using StGit. Use for rebasing patches to new Harbor versions, creating/modifying patches, resolving conflicts, or importing/exporting the patch series.
argument-hint: "[task]"
allowed-tools: Bash(stg *), Bash(git *), Read, Glob, Grep
---

# 8gcr-ee Patch Management with StGit

Maintains Enterprise Edition patches on top of OSS Harbor using StGit.

## Project-Specific Paths

| Item | Path |
|------|------|
| Patches directory | `8gcr-ee/patches/` |
| Series file | `8gcr-ee/patches/series` |
| Decision record | `8gcr-ee/decision-records/0001-managing-8gcr-modifications-on-harbor.md` |

## Project Commands

```bash
# Import all patches onto OSS Harbor
stg init && stg import -s 8gcr-ee/patches/series

# Export patches back to files (after changes)
stg export -d 8gcr-ee/patches/ -n
```

## Conventions

- **Naming:** `NNNN-short-description.patch` (4-digit prefix, e.g., `0001-`)
- **Ordering:** Schema migrations patch (`0012-schema-migrations`) always last
- **Location:** All patches in `8gcr-ee/patches/`, listed in `series` file

## Key Workflows

### Rebase to New OSS Harbor Version

```bash
stg pop -a                              # remove all patches
stg rebase v2.15.0                      # rebase to new version
stg push                                # apply one-by-one, resolve conflicts
stg export -d 8gcr-ee/patches/ -n       # export updated patches
```

### Modify Existing Patch

```bash
stg goto <patch-name>                   # navigate to patch
# make changes...
stg refresh                             # update patch
stg push -a                             # re-apply remaining
stg export -d 8gcr-ee/patches/ -n       # export
```

### Create New Patch

```bash
stg push -a                             # ensure all applied
stg new 0013-feature-name -m "feat: description"
# implement...
stg refresh
stg export -d 8gcr-ee/patches/ -n
echo "0013-feature-name.patch" >> 8gcr-ee/patches/series
```

## Series Status

`stg series` output: `>` current, `+` applied, `-` unapplied
