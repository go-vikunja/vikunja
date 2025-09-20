import type {Editor} from '@tiptap/core'
import inputPrompt from '@/helpers/inputPrompt'

export async function setLinkInEditor(pos: { x: number, y: number }, editor: Editor) {
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
