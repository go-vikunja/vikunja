import {Extension} from '@tiptap/core'
import {TextSelection} from '@tiptap/pm/state'
import type {EditorState} from '@tiptap/pm/state'

/**
 * Tiptap's `splitBlock` decides whether it may split *before* it deletes the selection. When the
 * deletion also removes the block the cursor was in - which ProseMirror does whenever the selection
 * starts at the very beginning of a block - the position it then splits at no longer sits inside a
 * block and the step throws `TransformError: Inserted content deeper than insertion position`.
 */
function splitAfterDeleteWouldLeaveDocumentLevelPosition(state: EditorState): boolean {
	const {selection} = state

	if (selection.empty || !(selection instanceof TextSelection)) {
		return false
	}

	const tr = state.tr
	tr.deleteSelection()

	return tr.doc.resolve(tr.mapping.map(selection.$from.pos)).depth === 0
}

/**
 * Takes the deletion out of Enter and runs it as its own step, so whatever handles Enter afterwards
 * starts from an empty selection at a position that can actually be split.
 */
export const DeleteSelectionBeforeEnter = Extension.create({
	name: 'deleteSelectionBeforeEnter',

	// Must run before the core keymap turns Enter into a splitBlock.
	priority: 1000,

	addKeyboardShortcuts() {
		return {
			Enter: () => {
				if (!splitAfterDeleteWouldLeaveDocumentLevelPosition(this.editor.state)) {
					return false
				}

				this.editor.commands.deleteSelection()

				// Not handled: the remaining Enter handlers now run against the shortened document.
				return false
			},
		}
	},
})
