---
description: Strict Issue-Driven Development Workflow
---

# Strict Issue-Driven Development Workflow

This workflow ensures that every code change is proposed, reviewed, tracked via GitHub issues, and verified before being merged.

## Phase 0: Pre-flight Synchronization

1. **Verify Completion**: Check that the previous ticket is closed and the branch is deleted.
2. **Pull Latest**: Sync the local main branch with the remote.

```bash
git checkout main && git pull origin main
```

3. **Clean Status**: Ensure `git status` reveals no uncommitted changes.

## Phase 1: Research & Proposal

1. **Analyze**: Use search tools to understand the task scope.
2. **Check Restrictions**: Verify if any target files are in the **Frozen Components List** inside [interlock.md](file:///.agent/rules/interlock.md).
3. **Draft Proposal**: Create an artifact containing the Objective, Proposed Changes, Impact Analysis, and Testing Plan.
4. **Human Gate**: Submit the proposal to the user. **Wait for explicit approval.**

## Phase 2: GitHub Ticketing

1. **Create Issue**: Once proposal is approved, use `mcp_github-mcp-server_issue_write` to create a new issue. **CRITICAL: Implementation cannot start without this step.**
2. **Reference ID**: Extract the Issue Number (e.g., `#123`).
3. **Create Branch**: Create a feature branch linked to the issue.

```bash
git checkout -b feature/issue-[number]-[description]
```

## Phase 3: Implementation

1. **Atomic Commits**: Write code and commit in logical chunks.
2. **Commit Format**: All commit messages must reference the issue.
   - Example: `feat(api): add endpoint for user profiles (issue #123)`
3. **Safety**: Do not edit frozen files without specific written override approval.

## Phase 4: Verification (QA)

1. **Test Execution**: Run existing tests and create new ones for the feature.
2. **Visual Verification**: If UI-related, use the browser subagent to verify the changes.
3. **Refinement**: Fix any bugs found during this phase. **CRITICAL: Verification must be complete before a PR is created.**

## Phase 5: Pull Request & Close

1. **Create PR**: Create a Pull Request with a clear description and Closes #123. **ONLY proceed if Phase 4 is complete.**
2. **User Review**: Present the final walkthrough and request PR review/approval.
3. **Final Gate**: After the USER merges the PR:
   - Close the GitHub Issue (if not auto-closed).
   - Delete the feature branch locally and remotely.
   - Return to Phase 0 for the next task.
