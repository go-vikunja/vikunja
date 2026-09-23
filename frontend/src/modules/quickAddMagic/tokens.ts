import {PREFIXES, PrefixMode} from './prefixes'
import {parseTaskTextWithMatches} from './quickAddMagic'
import type {ParsedTaskText} from './types'

export type MagicTokenType = 'label' | 'project' | 'priority' | 'assignee' | 'repeat' | 'date'

export interface MagicToken {
	type: MagicTokenType,
	from: number,
	to: number,
	value: string,
}

const escapeRegExp = (s: string): string => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

/**
 * Parses like parseTaskText and also returns where in `text` each piece of magic sits.
 * Tokens are located in the order the parser strips them; claimed ranges are blanked
 * so later steps can't match inside them.
 */
export const parseTaskTextWithTokens = (
	text: string,
	prefixesMode: PrefixMode = PrefixMode.Default,
	now: Date = new Date(),
): {parsed: ParsedTaskText, tokens: MagicToken[]} => {
	const {result: parsed, matches} = parseTaskTextWithMatches(text, prefixesMode, now)
	const prefixes = PREFIXES[prefixesMode]
	// Magic-only titles are created literally, see useQuickAddTask.
	if (prefixes === undefined || parsed.text === '') {
		return {parsed, tokens: []}
	}

	const tokens: MagicToken[] = []
	let masked = text

	const claim = (type: MagicTokenType, from: number, to: number, value: string) => {
		tokens.push({type, from, to, value})
		masked = masked.slice(0, from) + ' '.repeat(to - from) + masked.slice(to)
	}

	const claimPrefixed = (type: MagicTokenType, prefix: string, items: string[]) => {
		items.forEach(item => {
			const value = escapeRegExp(item)
			const pattern = new RegExp(`(^|\\s)${escapeRegExp(prefix)}('${value}'|"${value}"|${value})(?=\\s|$)`, 'gi')
			for (const match of masked.matchAll(pattern)) {
				const from = match.index + match[1].length
				claim(type, from, match.index + match[0].length, item)
			}
		})
	}

	claimPrefixed('label', prefixes.label, parsed.labels)
	claimPrefixed('project', prefixes.project, parsed.project !== null ? [parsed.project] : [])
	claimPrefixed('priority', prefixes.priority, parsed.priority !== null ? [String(parsed.priority)] : [])
	claimPrefixed('assignee', prefixes.assignee, parsed.assignees)

	if (matches.repeat !== null) {
		const index = masked.indexOf(matches.repeat)
		const repeat = matches.repeat.trimStart()
		if (index !== -1) {
			const from = index + matches.repeat.length - repeat.length
			claim('repeat', from, from + repeat.length, repeat)
		}
	}

	matches.date.map(s => s.trim()).filter(Boolean).forEach(removed => {
		for (const match of masked.matchAll(new RegExp(escapeRegExp(removed), 'gi'))) {
			claim('date', match.index, match.index + match[0].length, match[0])
		}
	})

	return {parsed, tokens: tokens.sort((a, b) => a.from - b.from)}
}
