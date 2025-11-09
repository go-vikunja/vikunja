import { VueRenderer } from '@tiptap/vue-3'
import { computePosition, flip, shift, offset, autoUpdate } from '@floating-ui/dom'
import type { Editor } from '@tiptap/core'

import MentionList from './MentionList.vue'
import ProjectUserService from '@/services/projectUsers'
import { fetchAvatarBlobUrl, getDisplayName } from '@/models/user'
import type { IUser } from '@/modelTypes/IUser'
import type { MentionNodeAttrs } from '@tiptap/extension-mention'
interface MentionItem extends MentionNodeAttrs {
	id: string
	label: string
	username: string
	avatarUrl: string
}

async function searchUsersForProject(projectId: number, query: string): Promise<MentionItem[]> {
	const projectUserService = new ProjectUserService()
	
	// Use server-side search with the 's' parameter
	const users = await projectUserService.getAll({ projectId }, { s: query }) as IUser[]

	// Fetch avatar URLs for all users
	const usersWithAvatars = await Promise.all(
		users.map(async (user) => {
			const avatarUrl = await fetchAvatarBlobUrl(user, 32)
			return {
				id: String(user.id),
				label: getDisplayName(user),
				username: user.username,
				avatarUrl: avatarUrl as string,
			}
		}),
	)

	return usersWithAvatars
}

export default function mentionSuggestionSetup(projectId: number) {
	let debounceTimer: ReturnType<typeof setTimeout> | null = null

	return {
		char: '@',

		items: async ({ query }: { query: string }): Promise<MentionItem[]> => {
			if (!projectId) {
				return []
			}

			// Clear existing timer
			if (debounceTimer) {
				clearTimeout(debounceTimer)
			}

			// Return a promise that resolves after debounce delay
			return new Promise((resolve) => {
				debounceTimer = setTimeout(async () => {
					try {
						// Use server-side search - the backend will handle searching by username and display name
						const users = await searchUsersForProject(projectId, query)

						// Limit results to avoid overwhelming the UI
						const limit = query ? 10 : 5
						resolve(users.slice(0, limit))
					} catch (error) {
						console.error('Failed to fetch users for mentions:', error)
						resolve([])
					}
				}, 300) // 300ms debounce delay
			})
		},

		render: () => {
			let component: VueRenderer
			let popupElement: HTMLElement | null = null
			let cleanupFloating: (() => void) | null = null

			const virtualReference = {
				getBoundingClientRect: () => ({
					width: 0,
					height: 0,
					x: 0,
					y: 0,
					top: 0,
					left: 0,
					right: 0,
					bottom: 0,
				} as DOMRect),
			}

			return {
				onStart: (props: {
					editor: Editor
					clientRect?: (() => DOMRect | null) | null
					items: MentionItem[]
					command: (item: MentionItem) => void
				}) => {
					component = new VueRenderer(MentionList, {
						props,
						editor: props.editor,
					})

					if (!props.clientRect) {
						return
					}

					// Create popup element
					popupElement = document.createElement('div')
					popupElement.style.position = 'absolute'
					popupElement.style.top = '0'
					popupElement.style.left = '0'
					popupElement.style.zIndex = '4700'
					popupElement.appendChild(component.element!)
					document.body.appendChild(popupElement)					// Update virtual reference
					const rect = props.clientRect()
					if (rect) {
						virtualReference.getBoundingClientRect = () => rect
						// Set up floating positioning
						const updatePosition = () => {
							computePosition(virtualReference, popupElement!, {
								placement: 'bottom-start',
								middleware: [
									offset(8),
									flip(),
									shift({ padding: 8 }),
								],
							}).then(({ x, y }) => {
								if (popupElement) {
									popupElement.style.left = `${x}px`
									popupElement.style.top = `${y}px`
								}
							})
						}

						updatePosition()
						cleanupFloating = autoUpdate(virtualReference, popupElement, updatePosition)
					}
				},

				onUpdate(props: {
					editor: Editor
					clientRect?: (() => DOMRect | null) | null
					items: MentionItem[]
					command: (item: MentionItem) => void
				}) {
					component.updateProps(props)

					if (!props.clientRect || !popupElement) {
						return
					}

					// Update virtual reference
					const rect = props.clientRect()
					if (rect) {
						virtualReference.getBoundingClientRect = () => rect
					}
				},

				onKeyDown(props: { event: KeyboardEvent }) {
					if (props.event.key === 'Escape') {
						if (props.event.isComposing) {
							return false
						}

						if (popupElement) {
							popupElement.style.display = 'none'
						}

						return true
					}

					return component.ref?.onKeyDown(props)
				},

				onExit() {
					if (cleanupFloating) {
						cleanupFloating()
					}
					if (popupElement) {
						document.body.removeChild(popupElement)
						popupElement = null
					}
					component.destroy()
				},
			}
		},
	}
}
