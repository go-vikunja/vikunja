import {ListKeymap} from '@tiptap/extension-list'
import {joinTextblockBackward} from '@tiptap/pm/commands'
import type {Editor} from '@tiptap/core'
import type {ResolvedPos} from '@tiptap/pm/model'

function findListItemDepth($from: ResolvedPos, itemNames: string[]): number | null {
	for (let depth = $from.depth; depth > 0; depth--) {
		if (itemNames.includes($from.node(depth).type.name)) {
			return depth
		}
	}

	return null
}

function joinIntoPreviousListItem(editor: Editor, itemNames: string[]): boolean {
	const {selection} = editor.state

	if (!selection.empty) {
		return false
	}

	const {$from} = selection
	const itemDepth = findListItemDepth($from, itemNames)

	if (itemDepth === null || $from.parentOffset !== 0) {
		return false
	}

	// The cursor must sit in the item's very first block, not in a later paragraph of it.
	for (let depth = itemDepth; depth < $from.depth; depth++) {
		if ($from.index(depth) !== 0) {
			return false
		}
	}

	// The first item of a list keeps the upstream behaviour: backspace lifts it out of the list.
	if ($from.index(itemDepth - 1) === 0) {
		return false
	}

	return editor.commands.command(({state, dispatch}) => joinTextblockBackward(state, dispatch))
}

/**
 * Backspace at the start of a list item merges it into the item above instead of lifting it out
 * of the list. Upstream ListKeymap always lifts, which splits the list in two and leaves a bare
 * paragraph wedged between the halves (#3480).
 */
export const ListKeymapWithJoin = ListKeymap.extend({
	addKeyboardShortcuts() {
		const parent = this.parent?.() ?? {}

		return {
			...parent,
			Backspace: (props) => {
				if (this.editor.commands.undoInputRule()) {
					return true
				}

				const itemNames = this.options.listTypes.map(({itemName}) => itemName)

				return joinIntoPreviousListItem(this.editor, itemNames)
					|| parent.Backspace?.(props)
					|| false
			},
		}
	},
})
