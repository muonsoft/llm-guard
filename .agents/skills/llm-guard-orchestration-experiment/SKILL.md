---
name: llm-guard-orchestration-experiment
description: Apply llm-guard milestone, evidence, and measurement rules on top of the shared Herdr-first change-orchestration workflow. Use for non-trivial implementation milestones in llm-guard.
---

# LLM Guard Orchestration Experiment

Apply this project-owned overlay after reading `change-orchestration`,
`task-delegation`, and, when selected as transport, `herdr`. The shared
orchestrator owns transport selection and recovery; this overlay adds only
llm-guard milestone boundaries, evidence, and experiment measurements.

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

## Use the shared transport policy

Use Herdr first when the explicit workflow runs under a healthy `HERDR_ENV=1`
session. Otherwise use the `cursor-executor` fallback after its `doctor` check.
Keep the transport fixed during a milestone. If Herdr loses control, follow the
shared recovery procedure: inspect the diff, establish remaining work, then
switch only at a clean milestone boundary. Never run both as concurrent writers.

Use one Composer session per milestone, not per entire change. Reuse it for the
primary packet and the consolidated correction cycle, then start a new session
for the next milestone after the reviewed green checkpoint. Default to one
correction packet; use a second only for newly revealed verification evidence.

Herdr bypasses the persistent `cursor-executor` job record and repository writer
lock. Compensate by recording baseline Git status, enforcing the single-writer
rule, and capturing changed paths and verification evidence in the experiment
log.

## Make Herdr runs recoverable

Treat Herdr as experimental in the current Codex environment. A smoke test
on 2026-08-12 with Herdr 0.8.0 completed two prompts in one Composer session, but
subsequent `herdr agent read`, `get`, and `list` processes stopped responding
before the server logged an API request; the control plane later recovered. Treat
this as an observed Codex/Herdr compatibility risk, not a confirmed general Herdr
defect.

For every Herdr packet, name a unique
`.agent-orchestration/results/<milestone>.md` path in `Return format`. Require
Composer to write it after focused verification with the outcome, changed paths,
commands and results, blockers, and a unique completion marker. Keep the result
ephemeral and ignored by Git; use it as the durable handoff when terminal output
is temporarily unavailable.

Record the agent name, pane ID, Cursor session ID, result path, and baseline Git
status as soon as `agent start` succeeds. Apply a host/tool execution deadline to
every Herdr control call even when the subcommand has no timeout option. After a
settled prompt, classify the milestone transport:

- **Healthy:** the prompt settled and bounded `agent get` plus `agent read`
  completed. Continue the review loop normally.
- **Degraded:** the prompt settled as `done` or `idle`, its result file exists,
  but a read-only control call exceeded its deadline. Review the result file and
  full Git diff, run verification outside Herdr, and make one bounded liveness
  retry after review. If no correction is needed and verification is green,
  accept the milestone but do not reuse that Composer session. If correction is
  needed, continue only after the control plane recovers.
- **Blocked:** prompt submission or waiting failed, completion is ambiguous, the
  result file is missing, or a required correction cannot be delivered after the
  single liveness retry. Stop the milestone and report the evidence; do not fall
  back within it.

Do not accept a milestone from `done` or the result file alone. Sol still owns
whole-diff review and independent verification. Do not hammer a degraded socket,
call `herdr server stop`, kill the main process, or close an unresponsive pane
blindly. Verify Git outside Herdr, report any pane that could not be closed, and
capture bounded server-log evidence when readable without mutation.

Before treating this compatibility issue as resolved, pass a smoke test that
runs two consecutive prompts in the same Composer session and then successfully
completes `agent get`, bounded `agent read`, and `agent list`.

## Review and correction loop

1. Let Composer run the packet's focused checks and write the declared result
   file.
2. Read the result file and entire milestone diff, then rerun focused checks as
   Sol.
3. Finish the review before responding to any one finding.
4. Put all related actionable findings in one correction packet.
5. Continue the same Composer session once when the control plane is healthy or
   has recovered. Use a second correction only when new verification evidence
   appears.
6. Run broad repository verification at the milestone boundary, then establish a
   reviewed green checkpoint.

## Measure the experiment

Record every milestone in `docs/orchestration_experiment.md`: transport, primary
jobs, correction jobs, changed files and lines, executor time, orchestration and
review time, final verification, defects found after the first review, Herdr
transport state, control-call timeouts, recovery time, and orphaned panes.

Compare Herdr and cursor-executor runs only across completed milestones using the
same milestone-scale slicing policy. Prefer Herdr only when it reduces
orchestration overhead without increasing blocked milestones, failed final
verification, or escaped review defects. Treat elapsed time and diff size as
retrospective evidence, not quotas for future slicing.
