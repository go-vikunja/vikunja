import {Extension} from '@tiptap/core'
import Suggestion from '@tiptap/suggestion'

// Copied and adjusted from https://github.com/ueberdosis/tiptap/tree/252acb32d27a0f9af14813eeed83d8a50059a43a/demos/src/Experiments/Commands/Vue

export default Extension.create({
	name: 'slash-menu-commands',

	addOptions() {
		return {
			suggestion: {
				char: '/',
				command: ({editor, range, props}) => {
					props.command({editor, range})
				},
			},
		}
	},

	addProseMirrorPlugins() {
		return [
			Suggestion({
				editor: this.editor,
				...this.options.suggestion,
			}),
		]
	},
})
