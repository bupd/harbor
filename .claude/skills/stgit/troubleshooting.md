# StGit Troubleshooting

## Conflict During Push

**Symptom:** `stg push` fails with merge conflicts

```bash
# Check what's conflicting
git status

# Edit files to resolve (look for <<<<<<< ======= >>>>>>> markers)

# Stage resolved files
git add <resolved-files>

# Update the patch
stg refresh

# Continue
stg push -a
```

## Patch Won't Apply Cleanly

**Symptom:** Patch fails even after resolving visible conflicts

```bash
# Try with merge detection
stg push --merged

# Or undo and retry manually
stg undo --hard
stg push
# Resolve conflicts step by step
```

## Lost Changes After Pop

**Symptom:** Accidentally popped patches with uncommitted work

```bash
# Check recent operations
stg log

# Undo the pop
stg undo
```

## Corrupted Stack State

**Symptom:** StGit commands fail with internal errors

```bash
# Try repair first
stg repair

# If repair fails, reinitialize
stg branch --cleanup
stg init
stg import -s 8gcr-ee/patches/series
```

## Empty Patch After Refresh

**Symptom:** Patch has no changes (upstream already has them)

```bash
# Remove empty patches from stack
stg clean
```

## Wrong Patch Modified

**Symptom:** Changes went into wrong patch

```bash
# Undo the refresh
stg undo

# Go to correct patch
stg goto <correct-patch>

# Refresh there
stg refresh

# Return to previous position
stg push -a
```

## Rebase Fails Midway

**Symptom:** `stg rebase` stopped with conflicts

```bash
# Check current state
stg series
git status

# Resolve conflicts in current patch
git add <files>
stg refresh

# Continue rebasing remaining patches
stg push -a
```

## Can't Delete Applied Patch

**Symptom:** `stg delete` refuses on applied patch

```bash
# Pop it first
stg pop <patch>

# Then delete
stg delete <patch>

# Re-apply remaining
stg push -a
```

## Export Creates Wrong Filenames

**Symptom:** Exported patches don't match naming convention

```bash
# Use -n for numbered output matching series format
stg export -d 8gcr-ee/patches/ -n

# Rename if needed
mv 8gcr-ee/patches/01-foo.patch 8gcr-ee/patches/0001-foo.patch
```

## Series File Out of Sync

**Symptom:** Series file doesn't match actual patches

```bash
# List current stack
stg series

# Regenerate series file
stg series --all > 8gcr-ee/patches/series.new

# Review and replace
cat 8gcr-ee/patches/series.new
mv 8gcr-ee/patches/series.new 8gcr-ee/patches/series
```
