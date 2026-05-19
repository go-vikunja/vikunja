import Blockquote from '@tiptap/extension-blockquote'

/**
 * Blockquote extension that preserves `data-comment-id` across parse/serialize.
 * Used as the canonical reply marker: a comment that quotes another comment
 * stores the referenced comment's id on the wrapping blockquote, so both the
 * backend (for implicit-mention notifications) and the frontend (for the
 * jump-to-original chevron) can find it without a separate schema field.
 */
export const BlockquoteWithCommentId = Blockquote.extend({
	addAttributes() {
		return {
			...this.parent?.(),
			commentId: {
				default: null,
				parseHTML: (element: HTMLElement) => {
					const raw = element.getAttribute('data-comment-id')
					if (raw === null) {
						return null
					}
					const id = Number(raw)
					if (!Number.isInteger(id) || id <= 0) {
						return null
					}
					return id
				},
				renderHTML: (attributes) => {
					if (attributes.commentId === null || attributes.commentId === undefined) {
						return {}
					}
					return {
						'data-comment-id': String(attributes.commentId),
					}
				},
			},
		}
	},
})
