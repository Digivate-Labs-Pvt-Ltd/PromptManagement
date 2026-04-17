---
trigger: always_on
---

# Development Interlock Rules

This file enforces the mandatory development workflow and critical safety restrictions for the `PromptManagement` project.

## 1. Mandatory Development Workflow

All feature development or bug fixes MUST follow this strict 9-step sequence:

1. **Proposal**: Create a detailed proposal for the change.
2. **Human Review**: Wait for explicit human approval of the proposal.
3. **Issue Creation**: Create a corresponding issue in the `ideal-invention` GitHub repository. THIS IS MANDATORY BEFORE ANY CODE CHANGES.
4. **Implementation**: Begin coding ONLY after the issue is created and a branch is checked out. No code should be written before a ticket exists.
5. **Initial Testing**: Validate the implementation (including manual/browser tests).
6. **Bug Fixing**: Resolve any issues identified during initial testing.
7. **Test Case Creation**: Write comprehensive unit or integration test cases to ensure no regressions.
8. **GitHub Push & PR**: Push the final, validated code and tests to the remote repository and create a Pull Request **ONLY after all tests have passed and been verified**.
9. **Approval & Closure**: Seek final approval and close the GitHub ticket ONLY after the PR is merged, all steps are complete, and tests pass.

## 2. Core Constraints (CRITICAL)

- **NEVER work on implementation without ticket creation.**
- **NEVER create a Pull Request without comprehensive testing/verification.**

## 3. Issue-Driven Development

- **Mandatory Tickets**: A GitHub ticket/issue is REQUIRED before creating or editing any program files.
- **Open Tickets Only**: All implementation work must be done against an **OPEN** issue. Use `list_issues` or `search_issues` to verify the state of a ticket before starting. If an issue is already CLOSED, you must not perform any further coding against it.
- **No Retroactive Tickets**: Creating a ticket _after_ the work is done is a strictly forbidden violation of this protocol. The ticket must exist and be in the OPEN state before a single line of project code is modified.
- **Serial Execution**: Only ONE ticket shall be worked on at a time. You must never proceed to a new ticket until the previous one is fully closed.

## 4. Frozen Components

- **Definition**: Certain modules or files may be marked as **FROZEN**.
- **Restriction**: Frozen components and modules can NEVER be edited or changed by the AI agent.
- **Override Procedure**: If resolving an issue requires modifying a frozen component, you MUST seek **explicit written approval** from the human user before touching the file.

## 5. Frozen Components List

_(Add files/directories here as they are identified)_

- (None currently marked)
