# `/16-eyes audit-diff` — analyze a diff or PR

Read-only. Never edits code. **The exact same engine as `/16-eyes audit`** (same
persisted lenses, same verify → adversarial-review → synthesis pipeline, same
guards) — the only difference is scope: each lens investigates a diff's changed
hunks instead of the whole repository. Use this for reviewing a PR before merge;
use full `/16-eyes audit` for an occasional deep sweep of everything, including
code nobody touched recently.

## When invoked

1. **Determine `today`** from your own session context, as `YYYY-MM-DD`.

2. **Determine the diff scope**, in this order:
   - If the user's invocation named an explicit base ref/branch, or a PR number, use
     that.
   - Else, if the `GITHUB_BASE_REF` environment variable is set (Bash: `echo
     "$GITHUB_BASE_REF"` — present in a GitHub Actions `pull_request` job), use it as
     the base branch.
   - Else, detect the repo's default branch (Bash: `git symbolic-ref
     refs/remotes/origin/HEAD` and strip the `refs/remotes/origin/` prefix; if that
     fails — common on a shallow/fresh clone with no tracking symref — fall back to
     `main`, then `master`, whichever exists as `origin/<name>`).
   - `base` = the merge-base of that branch and `HEAD` (Bash: `git merge-base
     origin/<default> HEAD`), not the branch tip itself — so the diff is exactly what
     the PR introduces, not everything the base branch has gained since the PR
     started.
   - **If a PR number was named and `gh` is available**, prefer `gh pr diff <n>` for
     the diff text and `gh pr view <n> --json number,title,url` for metadata — this
     works even without that PR's branch checked out locally (the common case in CI).

3. **Gather the diff** (via `Bash` — the Workflow tool has no filesystem/git access):
   - Changed file list: `git diff --name-status <base>...HEAD` (or `gh pr view
     --json files` in PR mode).
   - Full unified diff: `git diff <base>...HEAD` (or `gh pr diff <n>`).
   - Read `.16-eyes/config.json` if present and drop any changed file matching
     `excludePatterns` from what you gather — never send excluded paths to a lens.
   - **Size guard**: if the diff is large (rule of thumb: over ~2000 changed lines,
     or includes lockfiles/minified/vendored/generated-looking paths), drop the
     largest/most-generated files from what gets sent to lenses. **Log which ones and
     why** in your final summary — never silently claim full coverage of a diff you
     actually trimmed.

4. **Read `.16-eyes/config.json`** (`depth`, `adversarial.votesPerFinding`,
   `language`, `lensesPointer`, default `.16-eyes/lenses.json`). **Read the lenses
   file. If it doesn't exist, run the Auto-bootstrap flow from `init-flow.md` now**
   (identical to `audit-flow.md` step 4) — then re-read it. `audit-diff` never
   designs its own lenses; it only ever reuses the persisted set.

5. **Call the `Workflow` tool** with `script` set to the *exact* contents of the code
   block below, and `args: { today, base, head: 'HEAD', prNumber, changedFiles,
   diffText, profile, lenses, depth, votesPerFinding, language }` (`profile` is
   `{ languages, domain_summary }` from the lenses file, same as `audit-flow.md`).
   This IS the user's explicit opt-in to multi-agent orchestration.

6. **Take the workflow's return value** and write **both** files with `Write`:
   - Markdown: `<outputDir>/SECURITY_AUDIT_DIFF_<today>.md`.
   - JSON: `<outputDir>/SECURITY_AUDIT_DIFF_<today>.json`.
   - Same `<outputDir>` resolution and same-day collision rule (`-2`, `-3`, ...) as
     `audit-flow.md`.
   - Then write/overwrite `.16-eyes/last-diff-run.json` (path from
     `config.lastDiffRunPointer`, default `.16-eyes/last-diff-run.json`) with
     `{ markdown, json, generatedAt, base, head, prNumber, stats }` — parallel to
     `audit`'s `last-run.json`, so `/16-eyes fix` can find either report family.

7. **Report a short summary**: the diff range reviewed (`base..head`, or `PR #N`),
   any files dropped by the size guard, counts (findings confirmed real, safe vs
   risky, refuted/corrupted), and the paths to both files. **If step 4
   auto-bootstrapped config/lenses**, say so. Remind the user this only reviewed the
   diff — pre-existing code is `/16-eyes audit`'s job, not this command's.

## The workflow script

```js
export const meta = {
  name: '16-eyes-audit-diff',
  description:
    "Run a repo's persisted investigation lenses scoped to a diff, verify every finding, adversarially review high-impact ones, and produce a classified report.",
  phases: [{ title: 'Lenses' }, { title: 'Verification' }, { title: 'Adversarial review' }, { title: 'Synthesis' }],
}

// ── Schemas (identical to audit-flow.md — keep them in sync) ──────────────
const FINDINGS_SCHEMA = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          title: { type: 'string' },
          file: { type: 'string' },
          line: { type: 'number' },
          description: { type: 'string' },
          initial_impact: { type: 'string', enum: ['high', 'medium', 'low'] },
          initial_probability: { type: 'string', enum: ['high', 'medium', 'low'] },
        },
        required: ['title', 'file', 'description'],
      },
    },
  },
  required: ['findings'],
}

const VERDICT_SCHEMA = {
  type: 'object',
  properties: {
    is_real: { type: 'boolean' },
    impact: { type: 'string', enum: ['high', 'medium', 'low'] },
    probability: { type: 'string', enum: ['high', 'medium', 'low'] },
    fix_type: { type: 'string', enum: ['safe', 'risky'] },
    exploit_scenario: { type: 'string' },
    why: { type: 'string' },
    suggested_fix: { type: 'string' },
  },
  required: ['is_real', 'impact', 'probability', 'fix_type', 'why'],
}

const REFUTE_SCHEMA = {
  type: 'object',
  properties: { refuted: { type: 'boolean' }, reason: { type: 'string' } },
  required: ['refuted', 'reason'],
}

const EXEC_SUMMARY_SCHEMA = {
  type: 'object',
  properties: { summary: { type: 'string' } },
  required: ['summary'],
}

// ── i18n ───────────────────────────────────────────────────────────────────
const LANGUAGE_NAMES = { en: 'English', pt: 'Brazilian Portuguese', es: 'Spanish' }

const T = {
  en: {
    intro:
      '> Generated by the `16-eyes` skill: runs this repo\'s tailored investigation lenses (designed once by `/16-eyes init`) against a DIFF only — not the whole repository — verifies each finding skeptically, adversarially reviews (multiple independent skeptics) high-impact findings, and classifies the remainder into SAFE (mechanical fix) vs RISKY (needs a human decision). For a deep sweep of everything, including code this diff never touched, use `/16-eyes audit` instead.',
    scope: 'Scope',
    methodology: 'Methodology',
    lensesRun: 'Lenses run',
    rawToVerified: 'Raw → verified findings',
    corruptedSuffix: 'with corrupted/failed verification (see appendix)',
    adversarialReview: 'Adversarial review',
    adversarialSummary: (h, v, r) =>
      `${h} high-impact finding(s) went through ${v} independent reviewer(s) trying to refute them; ${r} were refuted and discarded.`,
    riskMatrixHeading: 'Risk matrix (impact × probability)',
    matrixHeaderCol: 'Impact \\ Probability',
    low: 'Low',
    medium: 'Medium',
    high: 'High',
    safeHeading: 'SAFE findings — mechanical fix, no behavior change',
    riskyHeading: 'RISKY findings — need a human decision before fixing',
    none: '_None._',
    appendixHeading: 'Appendix — discarded',
    falsePositiveHeading: 'False positive',
    adversarialDiscardedHeading: 'Discarded in adversarial review',
    corruptedHeading: 'Corrupted/failed verification — needs manual review',
    findingWhere: 'Where',
    findingLens: 'Lens',
    findingImpactProb: 'Impact/Probability',
    findingWhat: 'What',
    findingExploit: 'Exploit scenario',
    findingWhyRisky: "Why it's risky",
    findingWhySafe: "Why it's safe to fix",
    findingSuggestedFix: 'Suggested fix',
    notSpecified: '(not specified)',
    reviewersRefuted: 'reviewer(s) refuted',
  },
  pt: {
    intro:
      '> Gerado pelo skill `16-eyes`: roda as lentes de investigação deste repositório (desenhadas uma vez pelo `/16-eyes init`) contra um DIFF apenas — não o repositório inteiro — verifica cada achado ceticamente, faz revisão adversarial (múltiplos céticos independentes) nos achados de impacto alto, e classifica o resíduo em SAFE (correção mecânica) vs RISKY (decisão humana necessária). Para um sweep completo, incluindo código que este diff não tocou, use `/16-eyes audit`.',
    scope: 'Escopo',
    methodology: 'Metodologia',
    lensesRun: 'Lentes rodadas',
    rawToVerified: 'Achados brutos → verificados',
    corruptedSuffix: 'com verificação corrompida/falha (ver apêndice)',
    adversarialReview: 'Revisão adversarial',
    adversarialSummary: (h, v, r) =>
      `${h} achado(s) de impacto alto passaram por ${v} revisor(es) independente(s) tentando refutar; ${r} foram refutados e descartados.`,
    riskMatrixHeading: 'Matriz de risco (impacto × probabilidade)',
    matrixHeaderCol: 'Impacto \\ Probabilidade',
    low: 'Baixa',
    medium: 'Média',
    high: 'Alta',
    safeHeading: 'Achados SAFE — correção mecânica, sem mudança de comportamento',
    riskyHeading: 'Achados RISKY — precisam de decisão humana antes de corrigir',
    none: '_Nenhum._',
    appendixHeading: 'Apêndice — descartados',
    falsePositiveHeading: 'Falso-positivo',
    adversarialDiscardedHeading: 'Refutado na revisão adversarial',
    corruptedHeading: 'Verificação corrompida/falhou — revisar manualmente',
    findingWhere: 'Onde',
    findingLens: 'Lente',
    findingImpactProb: 'Impacto/Probabilidade',
    findingWhat: 'O quê',
    findingExploit: 'Cenário de exploração',
    findingWhyRisky: 'Por quê é risky',
    findingWhySafe: 'Por quê é seguro corrigir',
    findingSuggestedFix: 'Sugestão de correção',
    notSpecified: '(não especificada)',
    reviewersRefuted: 'revisor(es) refutaram',
  },
  es: {
    intro:
      '> Generado por el skill `16-eyes`: ejecuta las lentes de investigación de este repositorio (diseñadas una vez por `/16-eyes init`) contra un DIFF solamente — no todo el repositorio — verifica cada hallazgo de forma escéptica, hace una revisión adversarial (varios escépticos independientes) de los hallazgos de alto impacto, y clasifica el resto en SAFE (corrección mecánica) vs RISKY (requiere decisión humana). Para un barrido completo, incluyendo código que este diff no tocó, usa `/16-eyes audit`.',
    scope: 'Alcance',
    methodology: 'Metodología',
    lensesRun: 'Lentes ejecutadas',
    rawToVerified: 'Hallazgos brutos → verificados',
    corruptedSuffix: 'con verificación corrupta/fallida (ver apéndice)',
    adversarialReview: 'Revisión adversarial',
    adversarialSummary: (h, v, r) =>
      `${h} hallazgo(s) de alto impacto pasaron por ${v} revisor(es) independiente(s) intentando refutarlos; ${r} fueron refutados y descartados.`,
    riskMatrixHeading: 'Matriz de riesgo (impacto × probabilidad)',
    matrixHeaderCol: 'Impacto \\ Probabilidad',
    low: 'Baja',
    medium: 'Media',
    high: 'Alta',
    safeHeading: 'Hallazgos SAFE — corrección mecánica, sin cambio de comportamiento',
    riskyHeading: 'Hallazgos RISKY — requieren decisión humana antes de corregir',
    none: '_Ninguno._',
    appendixHeading: 'Apéndice — descartados',
    falsePositiveHeading: 'Falso positivo',
    adversarialDiscardedHeading: 'Descartado en la revisión adversarial',
    corruptedHeading: 'Verificación corrupta/fallida — revisar manualmente',
    findingWhere: 'Dónde',
    findingLens: 'Lente',
    findingImpactProb: 'Impacto/Probabilidad',
    findingWhat: 'Qué',
    findingExploit: 'Escenario de explotación',
    findingWhyRisky: 'Por qué es risky',
    findingWhySafe: 'Por qué es seguro corregir',
    findingSuggestedFix: 'Corrección sugerida',
    notSpecified: '(no especificada)',
    reviewersRefuted: 'revisor(es) refutaron',
  },
}

// ── Helpers (identical to audit-flow.md — keep them in sync) ─────────────
function looksCorrupted(obj) {
  if (!obj || typeof obj !== 'object') return false
  const strings = Object.values(obj).filter((v) => typeof v === 'string' && v.length > 0)
  return strings.length > 0 && strings.every((v) => v.trim().toLowerCase() === 'test')
}

function dedupKey(f) {
  const file = (f.file || '').trim().toLowerCase()
  const line = f.line == null ? '?' : String(f.line)
  return `${file}:${line}`
}

function verifyPrompt(f, scopeHint, languageName) {
  return `You are adversarially verifying a security finding raised by another agent during a diff-scoped review${scopeHint ? ` (${scopeHint})` : ''}.

Finding: "${f.title}"
File: ${f.file}${f.line ? `:${f.line}` : ''}
Description: ${f.description}
Initial impact guess: ${f.initial_impact ?? 'unknown'} · Initial probability guess: ${f.initial_probability ?? 'unknown'}

Re-read the ACTUAL code at that location (and its real callers/config) yourself — do not trust the description above at face value. Then decide:
- is_real: is this a genuine issue in the code as it exists today (not hypothetical, not already mitigated elsewhere, not dead code)?
- impact: high/medium/low if exploited (money movement, auth bypass, data breach = high; degraded UX or internal-only = low).
- probability: high/medium/low that this is actually reachable/exploitable given real callers and current config — not just "possible in theory".
- fix_type: "safe" if a fix is purely mechanical and changes no behavior for any legitimate flow today (e.g. add a missing check, pin a version); "risky" if fixing it changes behavior for a flow that might be relied on today, touches money/auth, or needs a product decision (threshold, tolerance, UX tradeoff).
- exploit_scenario: concrete inputs/steps that would trigger it (empty if not real).
- suggested_fix: a specific, minimal fix.
- why: your reasoning, referencing the real code you read.

Be skeptical. A finding that sounds scary in the abstract but isn't actually reachable, or is already handled by a check elsewhere, is NOT real — say so.

Write "why", "exploit_scenario", and "suggested_fix" in ${languageName}.`
}

function refutePrompt(f, languageName) {
  return `Try to REFUTE this security finding from a diff-scoped review (it already passed one verification pass and was classified impact=high — this is a 2nd, adversarial, independent check before it goes in a report someone will act on).

Finding: "${f.title}"
File: ${f.file}${f.line ? `:${f.line}` : ''}
Description: ${f.description}
Why it was judged real: ${f.verdict?.why ?? ''}
Exploit scenario claimed: ${f.verdict?.exploit_scenario ?? ''}

Read the actual code yourself. Look hard for a reason this ISN'T actually exploitable: a guard elsewhere, a precondition that can't occur, dead code, a framework default that already prevents it, or a misread of the code. If you genuinely cannot find a flaw in the finding after really trying, refuted=false. If you are UNCERTAIN either way, default to refuted=true — a finding has to survive skepticism to earn a spot in the report, not just avoid being disproven.

Write "reason" in ${languageName}.`
}

// ── Script body ────────────────────────────────────────────────────────────

const today = (args && args.today) || 'unknown-date'
const base = (args && args.base) || 'unknown-base'
const head = (args && args.head) || 'HEAD'
const prNumber = args && args.prNumber ? String(args.prNumber) : null
const changedFiles = (args && Array.isArray(args.changedFiles) && args.changedFiles) || []
const diffText = (args && args.diffText) || ''
const profile = (args && args.profile) || {}
const depth = args && args.depth === 'quick' ? 'quick' : 'thorough'
const configVotes =
  args && Number.isFinite(args.votesPerFinding) && args.votesPerFinding > 0 ? args.votesPerFinding : null
const language = args && T[args.language] ? args.language : 'en'
const L = T[language]
const languageName = LANGUAGE_NAMES[language]

const scopeLabel = prNumber ? `PR #${prNumber}` : `${base}..${head}`

const lenses = ((args && Array.isArray(args.lenses) && args.lenses) || []).filter((l) => l && l.prompt && l.name)
if (lenses.length === 0) {
  return {
    reportMarkdown: `# 16 Eyes — diff review — ${today}\n\nFailed: no investigation lenses were available (\`.16-eyes/lenses.json\` was empty, and auto-bootstrap did not produce any). Run \`/16-eyes init\` and try again.`,
    reportJson: JSON.stringify({ error: 'no-lenses-available' }, null, 2),
    stats: null,
  }
}

function diffLensPrompt(lens) {
  return `You are reviewing ONLY a diff (${scopeLabel}) as part of a security review — not the whole repository. Your focus area: "${lens.focus}".

Changed files in this diff:
${changedFiles.map((f) => `- ${f}`).join('\n') || '(none)'}

Unified diff:
\`\`\`diff
${diffText}
\`\`\`

${lens.prompt}

Investigate ONLY within these changed hunks, from your focus area above. You may use \`Read\` on the full current file content for necessary surrounding context (e.g. to see a function's full body, or where a caller comes from), but do not go hunting for unrelated issues elsewhere in the repository — that is full \`/16-eyes audit\`'s job, not this one. If nothing in this diff falls under your focus area, return \`{"findings": []}\` — an empty result is a normal, expected outcome for most lenses on most diffs.`
}

// ── Lenses → verification (pipeline, no barrier) ──────────────────────────
phase('Lenses')
const seenKeys = new Set()
const perLensVerified = await pipeline(
  lenses,
  (lens) => agent(diffLensPrompt(lens), { schema: FINDINGS_SCHEMA, phase: 'Lenses', label: `lens:${lens.name}`, model: 'sonnet' }),
  (raw, lens) => {
    const findings = (raw?.findings || []).filter((f) => f && f.title && f.file)
    const fresh = findings.filter((f) => {
      const k = dedupKey(f)
      if (seenKeys.has(k)) return false
      seenKeys.add(k)
      return true
    })
    if (fresh.length < findings.length) {
      log(`${lens.name}: ${findings.length - fresh.length} finding(s) dropped by dedup (same file:line already seen)`)
    }
    return parallel(
      fresh.map((f) => () =>
        agent(verifyPrompt(f, scopeLabel, languageName), {
          schema: VERDICT_SCHEMA,
          phase: 'Verification',
          label: `verify:${lens.name}`,
          model: 'sonnet',
        }).then((v) => {
          const corrupted = !v || looksCorrupted(v)
          return { ...f, lens: lens.name, verdict: corrupted ? null : v, verdict_corrupted: corrupted }
        }),
      ),
    )
  },
)
const allVerified = perLensVerified.flat().filter(Boolean)

const corrupted = allVerified.filter((f) => f.verdict_corrupted)
const needsVerdict = allVerified.filter((f) => !f.verdict_corrupted && f.verdict)
const realFindings = needsVerdict.filter((f) => f.verdict.is_real)
const falsePositives = needsVerdict.filter((f) => !f.verdict.is_real)
log(
  `${allVerified.length} unique finding(s) verified: ${realFindings.length} real, ${falsePositives.length} false-positive, ${corrupted.length} corrupted/failed verification (manual review)`,
)

// ── Adversarial review (high impact only) ────────────────────────────────
phase('Adversarial review')
const highImpact = realFindings.filter((f) => f.verdict.impact === 'high')
const otherImpact = realFindings.filter((f) => f.verdict.impact !== 'high')

const baseVotes = configVotes ?? (depth === 'quick' ? 1 : 3)
const votesPerFinding = budget.total && budget.remaining() < 200_000 ? 1 : baseVotes
if (highImpact.length > 0)
  log(`Adversarial review: ${highImpact.length} high-impact finding(s), ${votesPerFinding} refuter(s) each`)

const adversarial = await parallel(
  highImpact.map((f) => () =>
    parallel(
      Array.from({ length: votesPerFinding }, (_, i) => () =>
        agent(refutePrompt(f, languageName), {
          schema: REFUTE_SCHEMA,
          phase: 'Adversarial review',
          label: `refute:${f.lens}:${i}`,
          model: 'sonnet',
        }),
      ),
    ).then((votes) => {
      const valid = votes.filter(Boolean).filter((v) => !looksCorrupted(v))
      const refuteCount = valid.filter((v) => v.refuted).length
      const survived = valid.length > 0 && refuteCount * 2 < valid.length
      return { ...f, adversarial: { votes: valid.length, refuted: refuteCount, survived } }
    }),
  ),
)
const survivedHighImpact = adversarial.filter(Boolean).filter((f) => f.adversarial.survived)
const refutedHighImpact = adversarial.filter(Boolean).filter((f) => !f.adversarial.survived)
if (refutedHighImpact.length > 0) {
  log(
    `${refutedHighImpact.length} high-impact finding(s) refuted by a majority of reviewers — dropped from the final report (listed in the appendix)`,
  )
}

const finalReal = [...survivedHighImpact, ...otherImpact]
const safeFindings = finalReal.filter((f) => f.verdict.fix_type === 'safe')
const riskyFindings = finalReal.filter((f) => f.verdict.fix_type === 'risky')

// ── Synthesis ───────────────────────────────────────────────────────────────
phase('Synthesis')
const execSummaryOut = await agent(
  `Write a short (3-6 sentence) executive summary in ${languageName} for a diff-scoped security review (${scopeLabel}), given these results:
- ${lenses.length} investigation lenses run against this diff${profile?.domain_summary ? ` (repo: ${profile.domain_summary})` : ''}.
- ${allVerified.length} candidate findings verified; ${realFindings.length} confirmed real, ${falsePositives.length} discarded as false-positive.
- ${refutedHighImpact.length} high-impact finding(s) were refuted by adversarial review and dropped.
- ${safeFindings.length} findings are SAFE to fix mechanically (no behavior change); ${riskyFindings.length} are RISKY (need a product/human decision before fixing).
Do not list individual findings — just the shape of the result and what the reader should do next. Mention explicitly that this only reviewed the diff, not the whole repository.`,
  { schema: EXEC_SUMMARY_SCHEMA, phase: 'Synthesis', label: 'exec-summary', model: 'sonnet' },
)
const execSummary = execSummaryOut?.summary || ''

function withId(f, i) {
  return { id: `${f.lens}-${i}`, ...f }
}
const safeStructured = safeFindings.map(withId)
const riskyStructured = riskyFindings.map(withId)

function impactProbLabel(f) {
  return `${f.verdict.impact}/${f.verdict.probability}`
}

function findingBlock(f, i) {
  return `### ${i + 1}. ${f.title}

- **${L.findingWhere}:** \`${f.file}${f.line ? `:${f.line}` : ''}\`
- **${L.findingLens}:** ${f.lens} · **${L.findingImpactProb}:** ${impactProbLabel(f)}
- **${L.findingWhat}:** ${f.description}
${f.verdict.exploit_scenario ? `- **${L.findingExploit}:** ${f.verdict.exploit_scenario}\n` : ''}- **${f.verdict.fix_type === 'risky' ? L.findingWhyRisky : L.findingWhySafe}:** ${f.verdict.why}
- **${L.findingSuggestedFix}:** ${f.verdict.suggested_fix || L.notSpecified}
`
}

const riskMatrix = (() => {
  const cells = {
    high: { high: [], medium: [], low: [] },
    medium: { high: [], medium: [], low: [] },
    low: { high: [], medium: [], low: [] },
  }
  for (const f of finalReal) {
    const imp = cells[f.verdict.impact] ? f.verdict.impact : 'medium'
    const prob = cells[imp][f.verdict.probability] ? f.verdict.probability : 'medium'
    cells[imp][prob].push(f.title)
  }
  const row = (imp) =>
    `| **${L[imp]}** | ${cells[imp].low.join(' · ') || '—'} | ${cells[imp].medium.join(' · ') || '—'} | ${cells[imp].high.join(' · ') || '—'} |`
  return `| ${L.matrixHeaderCol} | ${L.low} | ${L.medium} | ${L.high} |\n|---|---|---|---|\n${row('high')}\n${row('medium')}\n${row('low')}`
})()

const corruptedSuffixText = corrupted.length ? `, ${corrupted.length} ${L.corruptedSuffix}` : ''

const reportMarkdown = `# 16 Eyes — diff review — ${today}

${L.intro}

- **${L.scope}:** ${scopeLabel}${prNumber ? ` (base \`${base}\` → head \`${head}\`)` : ''} · ${changedFiles.length} file(s) changed

${execSummary}

## ${L.methodology}

- **${L.lensesRun} (${lenses.length}):** ${lenses.map((l) => `\`${l.name}\` (${l.focus})`).join(', ')}
- **${L.rawToVerified}:** ${allVerified.length} candidates, ${realFindings.length} confirmed real, ${falsePositives.length} discarded as false positive${corruptedSuffixText}.
- **${L.adversarialReview}:** ${L.adversarialSummary(highImpact.length, votesPerFinding, refutedHighImpact.length)}

## ${L.riskMatrixHeading}

${riskMatrix}

---

## ${L.safeHeading} (${safeFindings.length})

${safeFindings.length === 0 ? L.none : safeFindings.map(findingBlock).join('\n')}

---

## ${L.riskyHeading} (${riskyFindings.length})

${riskyFindings.length === 0 ? L.none : riskyFindings.map(findingBlock).join('\n')}

---

## ${L.appendixHeading}

${falsePositives.length === 0 ? '' : `### ${L.falsePositiveHeading} (${falsePositives.length})\n\n${falsePositives.map((f) => `- **${f.title}** (\`${f.file}${f.line ? `:${f.line}` : ''}\`) — ${f.verdict.why}`).join('\n')}\n\n`}${refutedHighImpact.length === 0 ? '' : `### ${L.adversarialDiscardedHeading} (${refutedHighImpact.length})\n\n${refutedHighImpact.map((f) => `- **${f.title}** (\`${f.file}${f.line ? `:${f.line}` : ''}\`) — ${f.adversarial.refuted}/${f.adversarial.votes} ${L.reviewersRefuted}`).join('\n')}\n\n`}${corrupted.length === 0 ? '' : `### ${L.corruptedHeading} (${corrupted.length})\n\n${corrupted.map((f) => `- **${f.title}** (\`${f.file}${f.line ? `:${f.line}` : ''}\`)`).join('\n')}\n`}
`

const stats = {
  lenses: lenses.length,
  candidates: allVerified.length,
  real: realFindings.length,
  falsePositives: falsePositives.length,
  safe: safeFindings.length,
  risky: riskyFindings.length,
  refutedHighImpact: refutedHighImpact.length,
  corrupted: corrupted.length,
}

const reportJson = JSON.stringify(
  {
    schemaVersion: 1,
    date: today,
    diffScope: { base, head, prNumber, changedFiles },
    stats,
    findings: { safe: safeStructured, risky: riskyStructured },
    appendix: {
      falsePositives: falsePositives.map((f, i) => withId(f, i)),
      refutedHighImpact: refutedHighImpact.map((f, i) => withId(f, i)),
      corrupted: corrupted.map((f, i) => withId(f, i)),
    },
  },
  null,
  2,
)

return { reportMarkdown, reportJson, stats }
```

## Design notes (for maintainers)

- **Same engine, different scope.** Verify/adversarial-review/synthesis logic, every
  guard (`looksCorrupted`, the "0 valid votes ≠ survived" rule, dedup) is identical to
  `audit-flow.md` on purpose — keep the two in sync if either changes. The only real
  difference is `diffLensPrompt()`, which wraps each persisted lens's own prompt with
  the diff content and an instruction to stay inside it.
- **Diff gathering happens outside the Workflow script** (step 2-3 of "When invoked"),
  same reason as `audit-flow.md`: the tool has no filesystem/git access, so `git diff`/
  `gh pr diff` output has to be fetched by the orchestrating agent and passed in via
  `args`.
- **No lens-design, ever, in this flow** — unlike `audit`, which can fall back to
  redesigning lenses if truly none exist (guarded by auto-bootstrap), `audit-diff` has
  no "whole repo" context to profile from a diff alone. It always reuses whatever
  `init` (interactive or auto-bootstrapped) already produced.
- **Empty findings per lens are the expected common case** — most lenses' focus areas
  won't intersect most diffs. This is intentionally simple (run every lens, let
  irrelevant ones return fast) rather than a fragile pre-filter guessing relevance from
  file paths.
- **`last-diff-run.json`** is a separate pointer from `audit`'s `last-run.json` so
  `/16-eyes fix` (see `fix-flow.md`) can tell a full-audit report from a diff-scoped one
  and pick whichever is actually newest/relevant.
