import {Extension} from '@tiptap/core'
import type {Editor} from '@tiptap/vue-3'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {Decoration, DecorationSet} from '@tiptap/pm/view'
import type {Node as ProseMirrorNode} from '@tiptap/pm/model'

import {getLinkPreview, type ILinkPreview} from '@/services/linkPreview'
import './linkPreview.scss'

const linkPreviewPluginKey = new PluginKey('linkPreview')

function isExternalHttpUrl(href: string): boolean {
	try {
		const url = new URL(href, window.location.origin)
		return (url.protocol === 'http:' || url.protocol === 'https:') && url.host !== window.location.host
	} catch {
		return false
	}
}

function hostnameOf(href: string): string {
	try {
		return new URL(href).hostname
	} catch {
		return href
	}
}

// Build the card as plain DOM (text is set via textContent, so nothing from the
// remote page is ever parsed as HTML). Returned synchronously; the async
// preview fills it in once it resolves.
function buildCardElement(url: string): HTMLElement {
	const wrapper = document.createElement('div')
	wrapper.className = 'link-preview-widget'

	getLinkPreview(url).then((data) => {
		if (!data || (!data.title && !data.description)) {
			return
		}
		wrapper.appendChild(renderCard(url, data))
	})

	return wrapper
}

function renderCard(url: string, data: ILinkPreview): HTMLElement {
	const card = document.createElement('a')
	card.className = 'link-preview'
	card.href = data.url || url
	card.target = '_blank'
	card.rel = 'noopener noreferrer nofollow'

	if (data.image) {
		const image = document.createElement('img')
		image.className = 'link-preview__image'
		image.src = data.image
		image.alt = ''
		image.loading = 'lazy'
		image.addEventListener('error', () => image.remove())
		card.appendChild(image)
	}

	const body = document.createElement('div')
	body.className = 'link-preview__body'

	const site = document.createElement('div')
	site.className = 'link-preview__site'
	if (data.favicon) {
		const favicon = document.createElement('img')
		favicon.className = 'link-preview__favicon'
		favicon.src = data.favicon
		favicon.alt = ''
		favicon.loading = 'lazy'
		favicon.addEventListener('error', () => favicon.remove())
		site.appendChild(favicon)
	}
	site.appendChild(document.createTextNode(data.site_name || hostnameOf(url)))
	body.appendChild(site)

	if (data.title) {
		const title = document.createElement('div')
		title.className = 'link-preview__title'
		title.textContent = data.title
		body.appendChild(title)
	}
	if (data.description) {
		const description = document.createElement('div')
		description.className = 'link-preview__description'
		description.textContent = data.description
		body.appendChild(description)
	}

	card.appendChild(body)
	return card
}

// Walk the doc, and for every text block that contains external links, place a
// preview card widget right after the block. Each distinct URL gets one card.
function buildLinkPreviews(doc: ProseMirrorNode): DecorationSet {
	const decorations: Decoration[] = []
	const seen = new Set<string>()

	doc.descendants((node, pos) => {
		if (!node.isTextblock) {
			return true
		}

		const hrefs: string[] = []
		node.descendants((child) => {
			if (!child.isText) {
				return
			}
			child.marks.forEach((mark) => {
				const href = mark.attrs.href
				if (mark.type.name === 'link' && typeof href === 'string' && isExternalHttpUrl(href) && !seen.has(href)) {
					seen.add(href)
					hrefs.push(href)
				}
			})
		})

		if (hrefs.length === 0) {
			return false
		}

		const insertPos = pos + node.nodeSize
		hrefs.forEach((href) => {
			decorations.push(Decoration.widget(insertPos, () => buildCardElement(href), {
				side: 1,
				key: `link-preview-${href}`,
			}))
		})

		return false
	})

	return DecorationSet.create(doc, decorations)
}

// LinkPreviewExtension renders Slack-style preview cards for external links in a
// task description/comment, but only while the editor is read-only (preview
// mode). The cards are ProseMirror widget decorations, so they never become
// part of the stored document.
export const LinkPreviewExtension = Extension.create({
	name: 'linkPreview',

	addProseMirrorPlugins() {
		const editor = this.editor as Editor

		return [
			new Plugin<DecorationSet>({
				key: linkPreviewPluginKey,
				state: {
					init: (_, state) => editor.isEditable ? DecorationSet.empty : buildLinkPreviews(state.doc),
					apply(tr, value, _oldState, newState) {
						if (editor.isEditable) {
							return DecorationSet.empty
						}
						if (tr.docChanged || value === DecorationSet.empty) {
							return buildLinkPreviews(newState.doc)
						}
						return value.map(tr.mapping, tr.doc)
					},
				},
				props: {
					decorations(state) {
						return linkPreviewPluginKey.getState(state)
					},
				},
			}),
		]
	},
})
