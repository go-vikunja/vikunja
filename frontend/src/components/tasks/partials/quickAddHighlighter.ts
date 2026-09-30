import {Plugin, PluginKey} from '@tiptap/pm/state'
import {Decoration, DecorationSet} from '@tiptap/pm/view'
import type {Node} from '@tiptap/pm/model'

import {parseTaskTextWithTokens, PrefixMode, type MagicToken} from '@/modules/quickAddMagic'
import {getLabelByExactTitle} from '@/client/queries/labels'
import type {Label} from '@/client/generated'
import {getLabelColor} from '@/composables/useLabelStyles'
import {getTextColor} from '@/helpers/color/getTextColor'
import {formatDateLong} from '@/helpers/time/formatDate'

export interface QuickAddHighlightContext {
	mode: PrefixMode,
	labels: Label[],
	projectExists: (title: string) => boolean,
}

export const quickAddHighlighterKey = new PluginKey<DecorationSet>('quickAddHighlighter')

// Mirrors the indentation and bullet stripping in parseSubtasksViaIndention.
const LINE_PREFIX = /^\s*((\* |\+ |- )(\[ \] )?)?/

export function createQuickAddHighlighter(getContext: () => QuickAddHighlightContext) {
	return new Plugin({
		key: quickAddHighlighterKey,
		state: {
			init: (_, state) => decorateDocument(state.doc, getContext()),
			apply(tr, old) {
				if (!tr.docChanged && !tr.getMeta(quickAddHighlighterKey)) return old
				return decorateDocument(tr.doc, getContext())
			},
		},
		props: {
			decorations(state) {
				return this.getState(state)
			},
		},
	})
}

export function decorateDocument(doc: Node, context: QuickAddHighlightContext) {
	const decorations: Decoration[] = []

	// Each paragraph is one task line; +1 steps into the paragraph node.
	doc.forEach((line, offset) => {
		const text = line.textContent
		const prefixLength = LINE_PREFIX.exec(text)?.[0].length ?? 0
		const start = offset + 1 + prefixLength
		const {parsed, tokens} = parseTaskTextWithTokens(text.slice(prefixLength), context.mode)
		tokens.forEach(token => decorations.push(Decoration.inline(
			start + token.from,
			start + token.to,
			tokenAttributes(token, parsed.date, context),
		)))
	})

	return DecorationSet.create(doc, decorations)
}

function tokenAttributes(token: MagicToken, date: Date | null, context: QuickAddHighlightContext) {
	const attributes: Record<string, string> = {
		class: `quick-add-token is-${token.type}`,
		'data-token-type': token.type,
	}

	if (token.type === 'label') {
		const label = getLabelByExactTitle(context.labels, token.value)
		if (label) {
			const color = getLabelColor(label)
			attributes.style = `background-color: ${color}; color: ${getTextColor(color)};`
		} else {
			attributes.class += ' is-new'
		}
	}

	if (token.type === 'project' && !context.projectExists(token.value)) {
		attributes.class += ' is-unknown'
	}

	if (token.type === 'date' && date !== null) {
		attributes.title = formatDateLong(date)
	}

	return attributes
}
