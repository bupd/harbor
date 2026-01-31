# StGit Workflows for 8gcr-ee

## Initialize Fresh OSS Harbor with Patches

```bash
git checkout v2.14.0
stg init
stg import -s 8gcr-ee/patches/series
stg series  # verify
```

## Modify an Existing Patch

```bash
stg goto 0002-ldap-admin-group-filter  # navigate to patch
# make code changes...
stg refresh                             # capture changes
stg push -a                             # re-apply remaining
stg export -d 8gcr-ee/patches/ -n       # export to files
```

## Create a New Feature Patch

```bash
stg push -a                                    # ensure all applied
stg new 0013-my-feature -m "feat: description" # create patch
# implement feature...
stg refresh                                    # capture changes
stg export -d 8gcr-ee/patches/ -n              # export
echo "0013-my-feature.patch" >> 8gcr-ee/patches/series
```

## Rebase Patches to New OSS Version

```bash
stg pop -a                    # remove all patches
git fetch origin
stg rebase v2.15.0            # rebase to new version
stg push                      # apply one at a time

# On conflict:
# 1. Edit files to resolve (look for <<<<<<< markers)
# 2. git add <resolved-files>
# 3. stg refresh
# 4. stg push (continue)

stg export -d 8gcr-ee/patches/ -n  # export rebased patches
```

## Reorder Patches

```bash
stg float <patch>                  # move to top
stg sink --to <target> <patch>     # move to position
stg export -d 8gcr-ee/patches/ -n  # export
```

## Squash Multiple Patches

```bash
stg squash -n <new-name> -m "message" patch1 patch2 patch3
stg export -d 8gcr-ee/patches/ -n
```

## Split a Patch

```bash
stg goto <patch>
stg spill                      # move changes to working tree
stg new part1 -m "First part"
git add <specific-files>
stg refresh
stg new part2 -m "Second part"
git add .
stg refresh
stg delete <original-patch>
stg push -a
stg export -d 8gcr-ee/patches/ -n
```

## Commit Message Format

```
feat|fix|chore: short description

Longer description explaining:
- What the patch does
- Why it's needed
- Any dependencies
```
