---
name: "spec-kit"
description: "Runs Spec Kit workflows in Trae. Invoke when user mentions spec-kit, specify, /speckit.* commands, or wants constitution/specify/plan/tasks/implement/converge flow."
---

# Spec Kit for Trae

This skill adapts GitHub `spec-kit` to Trae as a workspace-local skill.

Use this skill when the user:

- mentions `spec-kit`, `specify`, or spec-driven development
- uses `constitution`, `specify`, `plan`, `tasks`, `implement`, `converge`, `clarify`, `analyze`, `checklist`, or `taskstoissues` in a Spec Kit context
- references `/speckit.*` commands and wants the same workflow inside Trae

## Important Limits

- This is a Trae skill adapter, not an official native Trae integration from the `spec-kit` project.
- Do not claim that Trae supports `/speckit.*` slash commands natively.
- In Trae, trigger this workflow from normal user requests or explicit mentions of `spec-kit`.

## Workspace Assumptions

Before acting, inspect the repo root for these resources and prefer them in this order:

1. `.specify/`
2. `.github/agents/speckit.*.agent.md`
3. `.github/prompts/speckit.*.prompt.md`
4. local `specify` CLI

If `.specify/` already exists, treat the repo as an initialized Spec Kit project and reuse its templates, scripts, memory, and manifests.

If `.specify/` does not exist:

1. Check whether `specify` is installed.
2. If the user asked to enable Spec Kit in the repo, initialize it first.
3. Prefer a standard Spec Kit init that installs `.specify/` assets for the repo.
4. Explain briefly that Trae recognition is provided by this skill layer.

## Intent Routing

Map the user's request to one Spec Kit stage:

- `constitution` -> project principles and governance
- `specify` -> feature specification
- `clarify` -> resolve ambiguous requirements
- `plan` -> implementation plan and design artifacts
- `tasks` -> actionable task breakdown
- `implement` -> execute the generated tasks
- `converge` -> compare implementation vs spec/plan/tasks and append remaining work
- `analyze`, `checklist`, `taskstoissues` -> optional supporting workflows

If the user names a specific `/speckit.*` command, route directly to that stage.

## Execution Rules

For the selected stage, first read the matching project instruction file if it exists:

- `.github/agents/speckit.constitution.agent.md`
- `.github/agents/speckit.specify.agent.md`
- `.github/agents/speckit.clarify.agent.md`
- `.github/agents/speckit.plan.agent.md`
- `.github/agents/speckit.tasks.agent.md`
- `.github/agents/speckit.implement.agent.md`
- `.github/agents/speckit.converge.agent.md`
- `.github/agents/speckit.analyze.agent.md`
- `.github/agents/speckit.checklist.agent.md`
- `.github/agents/speckit.taskstoissues.agent.md`

If the matching `.github/agents/...` file does not exist, fall back to the corresponding `.github/prompts/...` file.

If neither exists, execute the stage using the local `.specify/` templates and scripts.

## Stage Notes

### `constitution`

- Operate on `.specify/memory/constitution.md`.
- Keep related templates in sync when the constitution changes.

### `specify`

- Create one feature spec at a time.
- Reuse `.specify/templates/spec-template.md`.
- Persist the active feature directory so downstream stages can find it.

### `plan`

- Prefer running `.specify/scripts/bash/setup-plan.sh --json` from the repo root.
- Generate plan/design artifacts in the active feature directory.

### `tasks`

- Prefer running `.specify/scripts/bash/setup-tasks.sh --json` from the repo root.
- Generate `tasks.md` in strict checklist format with task IDs and file paths.

### `implement`

- Implement from `tasks.md` in dependency order.
- Preserve the contract defined by the spec, plan, research, and tasks artifacts.

### `converge`

- Compare the current codebase against the active spec artifacts.
- Append remaining work as new tasks instead of silently changing scope.

## Behavior Expectations

- Always mention the generated or updated artifact paths in the response.
- Prefer reusing existing Spec Kit files over inventing parallel formats.
- Respect existing extension hooks under `.specify/extensions.yml` when present.
- Ask the user only when a missing choice materially changes scope or behavior.
- Keep answers concise and practical.

## Example Triggers

- "用 spec-kit 给这个需求出 spec"
- "按 spec-kit 流程先做 plan"
- "帮我跑一下 tasks 阶段"
- "/speckit.plan 为这个功能生成技术方案"
