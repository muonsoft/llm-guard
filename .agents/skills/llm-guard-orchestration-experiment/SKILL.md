---
name: llm-guard-orchestration-experiment
description: Run llm-guard coding milestones as a controlled orchestration experiment with GPT-5.6 Sol as orchestrator and reviewer, Composer 2.5 as the sole code writer, and optional Herdr transport. Use when planning, delegating, reviewing, or measuring non-trivial implementation work in llm-guard during the experiment.
---

# LLM Guard Orchestration Experiment

Apply this project-owned overlay after reading `task-delegation`. When variant C
is selected, also read `herdr` and treat the installed CLI as the syntax authority.

## Preserve role boundaries

- Keep GPT-5.6 Sol responsible for exploration, execution mapping, task packets,
  whole-diff review, verification, and acceptance.
- Route all non-trivial product code to Composer 2.5. Keep only trivial process
  edits inline with the orchestrator.
- Allow only one writing Composer in a worktree. Parallelize read-only checks only.
- Keep OpenSpec and verification ledgers orchestrator-owned.

## Prefer milestone-scale primary jobs

Default one green semantic milestone to one primary Composer job. Include the
implementation, tests, fixtures, callers, and mechanical fallout needed to make
that milestone green.

Split a milestone only when every resulting job has all three properties:

1. an independently stated acceptance boundary;
2. an independently useful focused verification boundary;
3. a green result that is independently reviewable and shippable.

If any property is missing, merge the work. File count, layer count, checklist
count, and estimated duration are observations, not automatic split triggers. Do
not translate OpenSpec checklist items one-to-one into jobs.

Before delegating, write an execution map naming milestones, projected primary
jobs, dependencies, and broad verification boundaries. Reject adjacent jobs that
repeat most packet context or merely pass one contract across layers.

## Run one declared variant

- **A — baseline:** use the project `cursor-executor` with the shared
  `task-delegation` slicing defaults.
- **B — large slices:** use the same `cursor-executor`, but apply the milestone
  split test above. Use this as the default experiment variant.
- **C — Herdr:** apply the same milestone split test with one long-lived Composer
  2.5 agent in a sibling Herdr pane.

Never mix variants within a milestone. Never run variants concurrently in the
same worktree.

For A or B, run `cursor-executor.mjs doctor`, then use `start` for the primary
packet and `resume` for the consolidated review packet.

For C, require `HERDR_ENV=1` and successful `herdr agent list` from the
orchestrator environment. Create a sibling pane without stealing focus, preserve
the repository working directory, and start a uniquely named `cursor` agent
pinned through native arguments to `composer-2.5`. Submit only the pointer to the
repository-local task packet. If Herdr reports `blocked` or `unknown`, inspect
`agent get` and bounded `agent read` output before acting. A failed Herdr
preflight blocks variant C; do not silently switch variants.

For variant C only, this overlay overrides the shared Codex host adapter's
requirement to invoke `cursor-executor`; all other `task-delegation` policy remains
in force.

Variant C bypasses the persistent `cursor-executor` job record and repository
writer lock. Compensate by recording the baseline Git status, enforcing the
single-writer rule, and capturing changed paths and verification evidence in the
experiment log.

## Review and correction loop

1. Let Composer run the packet's focused checks and report evidence.
2. Read the entire milestone diff and rerun focused checks as Sol.
3. Finish the review before responding to any one finding.
4. Put all related actionable findings in one correction packet.
5. Continue the same Composer session once. Use a second correction only when
   new verification evidence appears.
6. Run broad repository verification at the milestone boundary, then establish a
   reviewed green checkpoint.

## Measure the experiment

Record every milestone in `docs/orchestration_experiment.md`: variant, primary
jobs, correction jobs, changed files and lines, executor time, orchestration and
review time, final verification, and defects found after the first review.

Prefer B or C over the baseline only when they reduce primary-job and
orchestration overhead without increasing failed final verification or escaped
review defects. Treat elapsed time and diff size as retrospective evidence, not
quotas for future slicing.
