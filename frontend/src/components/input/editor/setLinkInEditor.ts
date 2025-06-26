import inputPrompt from '@/helpers/inputPrompt'
import type {Editor} from '@tiptap/core'

export async function setLinkInEditor(pos: DOMRect, editor: Editor | null) {
	const previousUrl = editor?.getAttributes('link').href || ''
	const url = await inputPrompt(pos, previousUrl)

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
