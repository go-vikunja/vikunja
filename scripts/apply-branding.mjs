#!/usr/bin/env node
// apply-branding.mjs — build-time white-label transform for the engine.
//
// Rebrands the *user-facing* product name baked into the engine (the upstream
// default is "Vikunja") to this platform's product name. It is deliberately
// surgical: it only rewrites strings a human actually sees (i18n values, the
// PWA manifest, the browser-tab title fallback, the logo alt text) and never
// touches internal identifiers — Go module paths (code.vikunja.io), SCSS vars
// ($vikunja-font), BroadcastChannel / lock / CSS-class ids, the window
// .vikunjaDesktop API contract, DB defaults, data-dir paths, or upstream URLs.
// Renaming any of those would break the build or our ability to merge upstream
// security fixes.
//
// Per-tenant look (colors/logo/title/favicon) is handled separately at RUNTIME
// by frontend/index.html reading /branding.json — see branding/README.md. This
// script only bakes the platform DEFAULT product name into the shared image and
// into server-rendered emails (pkg/i18n), which runtime theming cannot reach.
//
// It is idempotent (re-running is a no-op) and safe to run after every
// `git merge upstream` or Crowdin re-import — that is the whole point: source
// stays close to upstream, branding is re-applied deterministically at build.
//
// Usage:
//   node scripts/apply-branding.mjs                 # write, name from config/env/default
//   node scripts/apply-branding.mjs --name "Acme"   # override product name
//   node scripts/apply-branding.mjs --check         # dry-run; exit 1 if changes needed (CI)
//
// Config resolution for the product name (first hit wins):
//   --name <str>  >  $PROJECTOS_BRAND_NAME  >  branding.config.json .productName  >  "ProjectOS"

import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join, relative } from 'node:path'
import { readdirSync } from 'node:fs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const engineRoot = join(scriptDir, '..')

// ---- args / config -------------------------------------------------------
const args = process.argv.slice(2)
const check = args.includes('--check')
const nameFlagIdx = args.indexOf('--name')
const nameFromFlag = nameFlagIdx !== -1 ? args[nameFlagIdx + 1] : undefined

function loadConfig() {
	const p = join(engineRoot, 'branding.config.json')
	if (!existsSync(p)) return {}
	try { return JSON.parse(readFileSync(p, 'utf8')) } catch { return {} }
}
const config = loadConfig()

const PRODUCT_NAME =
	nameFromFlag ||
	process.env.PROJECTOS_BRAND_NAME ||
	config.productName ||
	'ProjectOS'

// Optional link overrides (kept upstream unless explicitly configured, because
// they point at real external services and a wrong value is worse than Vikunja's).
const LINKS = config.links || {} // { contact, docs }

// ---- helpers -------------------------------------------------------------
function listJson(dir) {
	const abs = join(engineRoot, dir)
	if (!existsSync(abs)) return []
	return readdirSync(abs)
		.filter((f) => f.endsWith('.json'))
		.map((f) => join(dir, f))
}

// Rebrand the capitalized product word "Vikunja" -> PRODUCT_NAME, but never
// when it is part of the "Vikunja.io" URL host. Lowercase "vikunja" is left
// untouched everywhere (it only ever appears in URLs, emails, asset ids and
// JSON keys — never as a rendered product name).
const brandProductWord = (text) => text.replace(/Vikunja(?!\.io)/g, PRODUCT_NAME)

// ---- rules ---------------------------------------------------------------
// Each rule: { file, apply(text) => text }. Files missing on disk are skipped.
const rules = []

// 1. i18n — every rendered UI string + every server-rendered email, all langs.
for (const f of [...listJson('pkg/i18n/lang'), ...listJson('frontend/src/i18n/lang')]) {
	rules.push({ file: f, apply: brandProductWord })
}

// 2. PWA manifest (installed-app name). Targeted so env.VIKUNJA_* stays intact.
rules.push({
	file: 'frontend/vite.config.ts',
	apply: (t) =>
		t
			.replace(/(\bname:\s*')Vikunja(')/g, `$1${PRODUCT_NAME}$2`)
			.replace(/(\bshort_name:\s*')Vikunja(')/g, `$1${PRODUCT_NAME}$2`),
})

// 3. Logo alt text (accessibility / rendered on broken image).
rules.push({
	file: 'frontend/src/components/home/Logo.vue',
	apply: (t) => t.replace(/(\balt=")Vikunja(")/g, `$1${PRODUCT_NAME}$2`),
})

// 4. Browser-tab title fallback (shown before per-tenant branding.json loads).
for (const f of ['frontend/src/composables/useTitle.ts', 'frontend/src/helpers/setTitle.ts']) {
	rules.push({ file: f, apply: brandProductWord })
}

// 5. Optional: user-facing external links, only when configured.
if (LINKS.contact) {
	rules.push({
		file: 'frontend/src/components/misc/Error.vue',
		apply: (t) => t.replace('https://vikunja.io/contact/', LINKS.contact),
	})
}
if (LINKS.docs) {
	rules.push({
		file: 'frontend/src/components/misc/WebhookManager.vue',
		apply: (t) => t.replace('https://vikunja.io/docs/webhooks/', LINKS.docs),
	})
}

// ---- run -----------------------------------------------------------------
let changed = 0
let scanned = 0
for (const rule of rules) {
	const abs = join(engineRoot, rule.file)
	if (!existsSync(abs)) continue
	scanned++
	const before = readFileSync(abs, 'utf8')
	const after = rule.apply(before)
	if (after === before) continue
	changed++
	const rel = relative(engineRoot, abs)
	if (check) {
		console.log(`would rebrand: ${rel}`)
	} else {
		writeFileSync(abs, after)
		console.log(`rebranded: ${rel}`)
	}
}

console.log(
	`\n${check ? 'check' : 'apply'} complete — product name "${PRODUCT_NAME}", ` +
		`${scanned} files scanned, ${changed} ${check ? 'need rebranding' : 'rebranded'}.`,
)

if (check && changed > 0) {
	console.error('\n✗ Branding not applied. Run: node scripts/apply-branding.mjs')
	process.exit(1)
}
