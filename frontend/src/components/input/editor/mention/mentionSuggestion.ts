import { VueRenderer } from '@tiptap/vue-3'
import type { Editor } from '@tiptap/core'

import MentionList from './MentionList.vue'
import { getPopupContainer } from '../popupContainer'
import { createSuggestionPopup, type SuggestionPopup } from '../suggestionPopup'
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
	// @ts-expect-error - projectId is used for URL replacement but not part of IAbstract
	const users = await projectUserService.getAll({ projectId }, { s: query }) as IUser[]

	// Fetch avatar URLs for all users
	const usersWithAvatars = await Promise.all(
		users.map(async (user) => {
			const avatarUrl = await fetchAvatarBlobUrl(user, 32)
			return {
				id: user.username,
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
			let popup: SuggestionPopup | null = null

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

					const rect = props.clientRect?.()
					if (!rect) {
						return
					}

					popup = createSuggestionPopup(getPopupContainer(props.editor), component.element!, rect)
				},

				onUpdate(props: {
					editor: Editor
					clientRect?: (() => DOMRect | null) | null
					items: MentionItem[]
					command: (item: MentionItem) => void
				}) {
					component?.updateProps(props)

					const rect = props.clientRect?.()
					if (rect) {
						popup?.setReferenceRect(rect)
					}
				},

				onKeyDown(props: { event: KeyboardEvent }) {
					if (props.event.key === 'Escape') {
						if (props.event.isComposing) {
							return false
						}

						if (popup) {
							popup.element.style.display = 'none'
						}

						return true
					}

					return component?.ref?.onKeyDown(props)
				},

				onExit() {
					popup?.destroy()
					popup = null
					component?.destroy()
				},
			}
		},
	}
}
