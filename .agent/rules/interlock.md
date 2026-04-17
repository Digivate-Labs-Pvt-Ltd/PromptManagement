---
trigger: always_on
---

# Development Interlock Rules

This file enforces the mandatory development workflow and critical safety restrictions for the `PromptManagement` project.

## 1. Mandatory Development Workflow

All feature development or bug fixes MUST follow this strict 9-step sequence:

1. **Proposal**: Create a detailed proposal for the change.
2. **Human Review**: Wait for explicit human approval of the proposal.
3. **Issue Creation**: Create a corresponding issue in the `PromptManagement` GitHub repository. THIS IS MANDATORY BEFORE ANY CODE CHANGES.
4. **Branching**: Checkout a new branch for the specific issue (e.g., `feature/issue-name` or `infra/issue-name`).
5. **Implementation**: Begin coding ONLY after the issue is created and a branch is checked out.
6. **Testing**: Write comprehensive unit or integration test cases for all new logial blocks.
7. **Local Validation**: Ensure all tests (new and existing) pass locally.
8. **GitHub Push & PR**: Push the validated code and create a Pull Request (PR). Every ticket completion requires a PR.
9. **Approval & Closure**: Seek final approval and close the GitHub ticket ONLY after the PR is merged.

## 2. Core Constraints (CRITICAL)

- **NEVER work on implementation without ticket creation.**
- **NEVER create a Pull Request without comprehensive testing/verification.**
- **NEVER skip the PR process for any sub-issue or task.**
- **NEVER commit logic without accompanying test cases.**

## 3. Architectural Constraints (CRITICAL)

- **POST-Only API**: All service endpoints and handlers MUST use the `POST` method. Action-based routing (e.g., `/items/promote`) is the standard pattern.
- **Native Go**: Use `net/http` standard library for routing and handlers. Avoid external web frameworks.
- **pgx/v5**: Use `pgx/v5` for all database interactions.

## 4. Frozen Components

- **Definition**: Certain modules or files may be marked as **FROZEN**.
- **Restriction**: Frozen components and modules can NEVER be edited or changed by the AI agent.
- **Override Procedure**: If resolving an issue requires modifying a frozen component, you MUST seek **explicit written approval** from the human user before touching the file.

## 5. Frozen Components List

_(Add files/directories here as they are identified)_

- (None currently marked)
