import type {Editor} from '@tiptap/core'
import inputPrompt from '@/helpers/inputPrompt'
import {i18n} from '@/i18n'

export async function setLinkInEditor(pos: DOMRect, editor: Editor | null | undefined) {
	const previousUrl = editor?.getAttributes('link').href || ''
	const url = await inputPrompt(pos, i18n.global.t('input.editor.urlPlaceholder'), previousUrl, editor ?? undefined)

	// cancelled
	if (url === null) {
		return
	}

	// empty
	if (url === '') {
		editor
			?.chain()
			.focus()
			.extendMarkRange('link')
			.unsetLink()
			.run()

		return
	}

	// update link
	editor
		?.chain()
		.focus()
		.extendMarkRange('link')
		.setLink({href: url, target: '_blank'})
		.run()
}
