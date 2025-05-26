import {Plugin, PluginKey} from '@tiptap/pm/state'
import {EditorView} from '@tiptap/pm/view'

export function placeholder(text: string) {
	const update = (view: EditorView) => {
		if (view.state.doc.textContent) {
			view.dom.removeAttribute('data-placeholder')
		} else {
			view.dom.setAttribute('data-placeholder', text)
		}
	}

	return new Plugin({
		key: new PluginKey('placeholder'),
		view(view: EditorView) {
			update(view)

			return {update}
		},
	})
}
