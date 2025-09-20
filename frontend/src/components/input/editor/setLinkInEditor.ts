import type {Editor} from '@tiptap/core'
import inputPrompt from '@/helpers/inputPrompt'

export async function setLinkInEditor(pos: { x: number, y: number }, editor: Editor) {
	const previousUrl = editor?.getAttributes('link').href || ''
	const clientRect: ClientRect = {
		left: pos.x,
		top: pos.y,
		right: pos.x,
		bottom: pos.y,
		width: 0,
		height: 0,
		x: pos.x,
		y: pos.y,
		toJSON: () => ({ left: pos.x, top: pos.y, right: pos.x, bottom: pos.y, width: 0, height: 0, x: pos.x, y: pos.y }),
	}
	const url = await inputPrompt(clientRect, previousUrl)

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
