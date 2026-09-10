# Fork Manifesto

## Purpose

Build an AI-assisted personal task manager and assistant that reduces mental load, helps a person focus on what matters, and balances their commitments against their actual capacity.

Work and everyday life belong in the same personal world. Productivity means sustainable progress toward meaningful outcomes—not filling every available hour or maximizing the number of completed tasks.

This document records product direction, not a claim that the capabilities described below already exist. Development proceeds incrementally from the inherited Vikunja codebase.

## One person, one world

Each installation serves one person. Teams, shared workspaces, multi-user collaboration, and a multi-tenant hosted service are not product goals.

The product is driven by its maintainer's own use, but its public source must remain usable by others on their own systems. Personal routines, preferences, capacity, and assistant memory belong in private installation data, not hard-coded defaults.

Single-user does not mean unauthenticated. Personal data and server access still require protection.

## Human intent, bounded assistant autonomy

The person decides which work enters the system. The assistant does not create new commitments on their behalf. Breaking an accepted task into actionable steps is assistance, not permission to introduce unrelated obligations.

After intake, the assistant supports the entire lifecycle: clarification, decomposition, prioritization, planning, execution, review, postponement, completion, and deliberate disposal.

The assistant may adjust the working plan within boundaries set by the person. Changing commitments or deadlines, or deciding to abandon work, requires their approval. Decisions should be explainable and changes reversible where possible. When evidence or authority is insufficient, the assistant asks instead of assuming.

Discarding work that no longer matters is a valid outcome, not a failure.

## Capacity before calendar density

Plans should account for available time, energy, existing workload, and observed progress. Capacity is not a fixed productivity quota.

The assistant helps expose overload and trade-offs. It should adapt the plan to the person rather than pressure the person to satisfy an unrealistic plan.

## Learn through evidence and retrospectives

Collect data that improves decisions: estimates versus outcomes, progress, recurring friction, and the person's own feedback. The exact data model will evolve through use rather than being fixed in advance.

Periodic retrospectives should lead to practical improvements in planning and working habits—not merely produce reports. Observations are evidence to discuss, not unquestionable judgments about the person.

Add data and capabilities when they improve decision quality. Remove what does not justify its complexity or privacy cost. More tracking is not inherently better.

## Borrow methods, not dogma

Use helpful ideas from Getting Things Done and other approaches, including reliable capture, clear next actions, and regular review. No methodology should become a rigid workflow the person must serve.

The test is whether an approach reduces friction and supports meaningful, sustainable progress in real use.

## Server-first, across devices

Build the server-side foundation first. Web, mobile, and desktop are intended ways to access the same personal world.

This is a direction, not a promise to deliver all clients at once. Client technologies and offline behavior will be decided when their implementation is in scope.

Self-hosting is part of the product: installation, configuration, updates, backup, and recovery should have clear, repeatable instructions. Ease of setup must not depend on the maintainer's private environment.

## Configurable external AI

Use an external AI service, initially through an OpenAI-compatible API contract. The installation owner chooses a compatible endpoint, model, and credentials.

A configurable endpoint does not imply compatibility with every provider's API. Add other contracts only when concrete needs justify them. Local model deployment, runtime management, and dedicated local-AI support are outside the current scope.

AI requests are coordinated server-side. Credentials belong in private configuration, never in public source or client bundles. The owner must be able to understand which personal data is sent to their chosen provider. Send only context needed for the operation; unrelated secrets must not enter model context.

Choosing an external service means relevant personal data can leave the installation. Self-hosting the application is not a claim that all AI processing stays local.

## Public code, private life

The public repository documents the product, not the maintainer's life.

Do not commit real tasks, personal schedules, assistant memories, private retrospective records, credentials, database dumps, or logs containing personal information. Examples, fixtures, screenshots, and documentation must use synthetic or properly sanitized data. Review diffs and staged content for accidental disclosure before publishing.

Personalization must remain separate from distributable code. Product decisions should be explained with generalizable reasoning, without exposing the private events that motivated them.

## Independent from Vikunja, respectful of its foundation

Vikunja is the starting point, not a compatibility target or a product roadmap we must follow.

Select upstream security fixes, bug fixes, and useful improvements deliberately. We do not promise routine full merges, feature parity, or permanent compatibility with upstream internals.

An inherited feature is not entitled to remain solely because it already exists. Add, reshape, or remove capabilities according to this manifesto, with explicit consideration of consequences for existing data and workflows.

Independence does not remove upstream license, attribution, or notice obligations. Preserve applicable obligations when modifying and distributing the code. This manifesto does not change the repository's licenses.

## Build in small, useful steps

Prefer changes that can be used, evaluated, and refined. Avoid building a general-purpose assistant platform or speculative configuration system before personal use demonstrates a need.

Evaluate progress through reduced mental load, realistic planning, useful assistance, and sustainable outcomes—not feature count alone.

## Maintain the divergence history

For every deliberate product-level addition, removal, or behavioral change relative to Vikunja, update the history below in the same change. Include architectural decisions that materially shape the fork.

This history explains the fork; it is not a duplicate of every commit or dependency update. Record routine maintenance only when it changes product direction or creates a meaningful divergence.

Each entry records:

- **Date and title**
- **Status:** decided, implemented, reverted, or superseded
- **What changed**
- **Why it changed**
- **Consequences and trade-offs**, including data or migration effects where relevant
- **References:** relevant files, commits, or pull requests when available

Keep past decisions readable. If direction changes, append a new entry referencing the earlier decision rather than silently rewriting history. Clearly distinguish intended behavior from shipped functionality. Never invent historical changes or references.

## Divergence history

### 2026-09-11 — Establish the independent personal-assistant direction

**Status:** Decided; the manifesto is documented, while the product changes described here remain incremental future work.

**What changed:** Adopted an independent, single-person product direction for this Vikunja fork. Defined AI assistance across the task lifecycle, bounded autonomy, capacity-aware planning, retrospective-driven improvement, server-first development, and intended web, mobile, and desktop access. Chose configurable external AI through an initial OpenAI-compatible API contract; dedicated local-AI support is outside current scope. Established public-code/private-data boundaries and a divergence-history convention.

**Why:** A personal assistant must help one person manage both work and everyday life, reduce mental load, and balance chosen commitments with actual capacity. That goal should guide the fork independently of upstream collaboration-oriented capabilities.

**Consequences and trade-offs:** Product decisions prioritize concrete personal use over broad market requirements. Others should still be able to self-host without inheriting the maintainer's personal configuration. Upstream improvements will be adopted selectively; divergence brings independent maintenance responsibility. External AI introduces provider availability, cost, and privacy considerations. This decision does not itself remove inherited multi-user features, add assistant capabilities, or deliver new clients.

**References:** This manifesto.
