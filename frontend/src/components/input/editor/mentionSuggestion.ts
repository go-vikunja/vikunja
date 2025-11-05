import {VueRenderer} from '@tiptap/vue-3'
import {computePosition, flip, shift, offset, autoUpdate} from '@floating-ui/dom'
import type {Editor} from '@tiptap/core'

import MentionList from './MentionList.vue'
import ProjectUserService from '@/services/projectUsers'
import {fetchAvatarBlobUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

interface MentionItem {
	id: number
	label: string
	username: string
	avatarUrl: string
}

let cachedUsers: MentionItem[] = []
let cacheProjectId: number | null = null

async function fetchUsersForProject(projectId: number): Promise<MentionItem[]> {
	// Return cached users if we're asking for the same project
	if (cacheProjectId === projectId && cachedUsers.length > 0) {
		return cachedUsers
	}

	const projectUserService = new ProjectUserService()
	const users = await projectUserService.getAll({projectId}, {}) as IUser[]
	
	// Fetch avatar URLs for all users
	const usersWithAvatars = await Promise.all(
		users.map(async (user) => {
			const avatarUrl = await fetchAvatarBlobUrl(user, 32)
			return {
				id: user.id,
				label: getDisplayName(user),
				username: user.username,
				avatarUrl: avatarUrl as string,
			}
		}),
	)

	// Cache the results
	cachedUsers = usersWithAvatars
	cacheProjectId = projectId

	return usersWithAvatars
}

export default function mentionSuggestionSetup(projectId: number) {
	return {
		char: '@',
		
		items: async ({query}: {query: string}): Promise<MentionItem[]> => {
			if (!projectId) {
				return []
			}

			try {
				const users = await fetchUsersForProject(projectId)
				
				if (!query) {
					return users.slice(0, 5) // Limit to 5 users when no query
				}

				// Filter users by query (search in both display name and username)
				return users
					.filter(user => 
						user.label.toLowerCase().includes(query.toLowerCase()) ||
						user.username.toLowerCase().includes(query.toLowerCase()),
					)
					.slice(0, 10) // Limit to 10 results
			} catch (error) {
				console.error('Failed to fetch users for mentions:', error)
				return []
			}
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
					clientRect: () => DOMRect
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
					popupElement.style.zIndex = '1000'
					popupElement.appendChild(component.element!)
					document.body.appendChild(popupElement)

					// Update virtual reference
					const rect = props.clientRect()
					virtualReference.getBoundingClientRect = () => rect

					// Set up floating positioning
					const updatePosition = () => {
						computePosition(virtualReference, popupElement!, {
							placement: 'bottom-start',
							middleware: [
								offset(8),
								flip(),
								shift({padding: 8}),
							],
						}).then(({x, y}) => {
							if (popupElement) {
								popupElement.style.left = `${x}px`
								popupElement.style.top = `${y}px`
							}
						})
					}

					updatePosition()
					cleanupFloating = autoUpdate(virtualReference, popupElement, updatePosition)
				},

				onUpdate(props: {
					editor: Editor
					clientRect: () => DOMRect
					items: MentionItem[]
					command: (item: MentionItem) => void
				}) {
					component.updateProps(props)

					if (!props.clientRect || !popupElement) {
						return
					}

					// Update virtual reference
					const rect = props.clientRect()
					virtualReference.getBoundingClientRect = () => rect
				},

				onKeyDown(props: {event: KeyboardEvent}) {
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
