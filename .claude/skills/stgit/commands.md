# StGit Command Reference

## Stack Inspection

| Command | Description |
|---------|-------------|
| `stg series` | List all patches with status |
| `stg series -d` | Include patch descriptions |
| `stg top` | Show topmost applied patch |
| `stg next` | Show next unapplied patch |
| `stg prev` | Show patch below current |
| `stg show [patch]` | Show patch diff |
| `stg files [patch]` | List files in patch |
| `stg patches <file>` | Show patches affecting file |

## Stack Navigation

| Command | Description |
|---------|-------------|
| `stg push` | Apply next unapplied patch |
| `stg push <patch>` | Apply up to specific patch |
| `stg push -n <N>` | Apply N patches |
| `stg push -a` | Apply all unapplied patches |
| `stg pop` | Remove topmost patch |
| `stg pop -n <N>` | Remove N patches |
| `stg pop -a` | Remove all applied patches |
| `stg goto <patch>` | Go to specific patch |
| `stg float <patch>` | Move patch to top |
| `stg sink <patch>` | Move patch toward bottom |

## Patch Modification

| Command | Description |
|---------|-------------|
| `stg new <name> -m "msg"` | Create new patch |
| `stg refresh` | Update current patch with changes |
| `stg refresh -p <patch>` | Update specific patch |
| `stg edit [patch]` | Edit patch message |
| `stg rename <old> <new>` | Rename patch |
| `stg delete <patch>` | Delete patch |
| `stg squash <patches>` | Combine patches |
| `stg spill` | Move changes to working tree |
| `stg fold <diff>` | Fold diff into current patch |

## Import/Export

| Command | Description |
|---------|-------------|
| `stg import -s <series>` | Import from series file |
| `stg import <mbox>` | Import from mailbox |
| `stg import --mail <file>` | Import email patch |
| `stg export -d <dir>` | Export patches to directory |
| `stg export -d <dir> -n` | Export with numbered filenames |

## Branch Operations

| Command | Description |
|---------|-------------|
| `stg init` | Initialize StGit on branch |
| `stg branch --list` | List StGit branches |
| `stg branch --cleanup` | Remove StGit metadata |
| `stg pull` | Pull and rebase stack |
| `stg rebase <target>` | Rebase stack to target |

## History & Undo

| Command | Description |
|---------|-------------|
| `stg log` | Show patch modification history |
| `stg log <patch>` | History for specific patch |
| `stg undo` | Undo last operation |
| `stg undo --hard` | Undo including working tree |
| `stg redo` | Redo undone operation |
| `stg reset` | Reset stack to earlier state |

## Cleanup

| Command | Description |
|---------|-------------|
| `stg clean` | Remove empty patches |
| `stg repair` | Repair corrupted stack |
