import {Node, mergeAttributes} from '@tiptap/core'
import {VueNodeViewRenderer} from '@tiptap/vue-3'
import MermaidBlock from './MermaidBlock.vue'

export const MermaidExtension = Node.create({
	name: 'mermaid',

	group: 'block',

	content: 'text*',

	code: true,

	defining: true,

	addAttributes() {
		return {
			language: {
				default: 'mermaid',
			},
		}
	},

	parseHTML() {
		return [
			{
				tag: 'pre[data-type="mermaid"]',
			},
		]
	},

	renderHTML({HTMLAttributes}) {
		return [
			'pre',
			mergeAttributes(HTMLAttributes, {'data-type': 'mermaid'}),
			['code', {}, 0],
		]
	},

	addNodeView() {
		return VueNodeViewRenderer(MermaidBlock)
	},

	addKeyboardShortcuts() {
		return {
			'Mod-Alt-m': () => this.editor.commands.setNode(this.name),
		}
	},
})
