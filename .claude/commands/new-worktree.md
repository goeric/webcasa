<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Set up a new git worktree for unrelated work. All work that does not belong
to the current branch must go in a separate worktree.

Ask the user for a short descriptive name if one wasn't provided (e.g.
"251-llm-timeouts", "fix-dashboard-crash").

Steps:

1. `git fetch origin`
2. `git worktree add ~/src/agent-work/<name> -b <branch-name> origin/main`
   - Branch name: conventional-commit style (e.g. `feat/251-llm-timeouts`,
     `fix/dashboard-crash`, `docs/scale-note`)
3. Set your working directory to `~/src/agent-work/<name>` for all subsequent
   commands
4. `direnv allow`
5. `direnv reload`

After these steps, confirm the worktree is ready and proceed with the task.
