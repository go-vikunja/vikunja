import {Node} from '@tiptap/core'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {VueNodeViewRenderer} from '@tiptap/vue-3'

import {parseTaskIdFromUrl} from '@/helpers/parseTaskIdFromUrl'

import TaskLinkView from './TaskLinkView.vue'

const PLAIN_ANCHOR_ATTRIBUTES = ['href', 'target', 'rel']

// Pasted taskLink attrs must serialize identically to the Link mark's anchors (NonInclusiveLink.configure in TipTap.vue), hence one shared constant.
export const LINK_HTML_ATTRIBUTES = {target: '_blank', rel: 'noopener noreferrer nofollow'}

// In-memory only: a plain <a href> whose text equals its same-origin /tasks/:id href
// is upgraded on parse and serialized back to the same anchor, so stored HTML stays plain.
export const TaskLink = Node.create({
	name: 'taskLink',

	// Plugins run in reverse declaration order, so this has to outrank TipTap.vue's
	// markdown paste handler, which would run a pasted url through marked instead.
	// Stays below the Link mark (1000) so pasting over a selection still links it.
	priority: 200,

	group: 'inline',
	inline: true,
	atom: true,
	selectable: true,
	draggable: false,

	addAttributes() {
		// Same order as the Link mark renders them, so re-saving is byte-stable.
		return {
			target: {
				default: null,
			},
			rel: {
				default: null,
			},
			href: {
				default: null,
			},
		}
	},

	parseHTML() {
		return [
			{
				tag: 'a[href]',
				// Must beat the Link mark rule (equal priority = marks first).
				priority: 100,
				getAttrs: (element) => {
					const href = element.getAttribute('href') ?? ''
					if (parseTaskIdFromUrl(href) === null) {
						return false
					}
					if (element.textContent?.trim() !== href) {
						return false
					}
					// Child markup and extra attributes would be dropped on re-render.
					if (element.children.length > 0) {
						return false
					}
					if (Array.from(element.attributes).some(attribute => !PLAIN_ANCHOR_ATTRIBUTES.includes(attribute.name))) {
						return false
					}
					return {
						target: element.getAttribute('target'),
						rel: element.getAttribute('rel'),
						href,
					}
				},
			},
		]
	},

	renderHTML({node, HTMLAttributes}) {
		return ['a', HTMLAttributes, node.attrs.href]
	},

	renderText({node}) {
		return node.attrs.href
	},

	addNodeView() {
		return VueNodeViewRenderer(TaskLinkView)
	},

	addProseMirrorPlugins() {
		return [
			new Plugin({
				key: new PluginKey('taskLinkPaste'),
				props: {
					handlePaste: (view, _event, slice) => {
						const {selection} = view.state
						if (!selection.empty || selection.$from.parent.type.spec.code) {
							return false
						}
						const text = slice.content.textBetween(0, slice.content.size, ' ').trim()
						if (/\s/.test(text) || parseTaskIdFromUrl(text) === null) {
							return false
						}
						return this.editor.commands.insertContent({
							type: this.name,
							attrs: {href: text, ...LINK_HTML_ATTRIBUTES},
						})
					},
				},
			}),
		]
	},
})
