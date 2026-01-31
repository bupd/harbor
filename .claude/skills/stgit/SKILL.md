---
name: stgit
description: Manage patch stacks with StGit (Stacked Git). Use when working with 8gcr-ee patches, rebasing patches to new OSS Harbor versions, creating new feature patches, modifying existing patches, resolving patch conflicts, or exporting/importing patch series.
argument-hint: "[command] [patch-name]"
allowed-tools: Bash(stg *), Bash(git *), Read, Glob, Grep
---

# StGit Patch Management

StGit manages a stack of patches on top of a Git branch for 8gcr-ee Enterprise Edition features on OSS Harbor.

## Essential Commands

| Task | Command |
|------|---------|
| Initialize | `stg init` |
| Import patches | `stg import -s 8gcr-ee/patches/series` |
| List stack | `stg series` |
| Apply all | `stg push -a` |
| Remove all | `stg pop -a` |
| Update patch | `stg refresh` |
| Export patches | `stg export -d 8gcr-ee/patches/ -n` |

## Series Status Symbols

- `>` Current (topmost) patch
- `+` Applied patch
- `-` Unapplied patch

## Additional Resources

Load these files based on the task at hand:

| Need | File |
|------|------|
| Step-by-step workflows (rebase, create, modify patches) | [workflows.md](workflows.md) |
| Full command reference | [commands.md](commands.md) |
| Fixing errors and conflicts | [troubleshooting.md](troubleshooting.md) |

## Project Conventions

- Patches location: `8gcr-ee/patches/`
- Series file: `8gcr-ee/patches/series`
- Naming: `NNNN-short-description.patch` (4-digit prefix)
- Schema migrations patch always last in series
