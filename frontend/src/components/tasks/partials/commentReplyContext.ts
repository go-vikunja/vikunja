import type {InjectionKey} from 'vue'
import type {ITaskComment} from '@/modelTypes/ITaskComment'

export interface CommentReplyContext {
	findComment: (id: number) => ITaskComment | undefined
	scrollToComment: (id: number) => void
}

export const commentReplyContextKey: InjectionKey<CommentReplyContext> = Symbol('commentReplyContext')

const HIGHLIGHT_CLASS = 'comment-highlight'
const HIGHLIGHT_DURATION_MS = 1500

export function scrollAndHighlightComment(id: number): void {
	const el = document.getElementById(`comment-${id}`)
	if (!el) {
		return
	}
	el.scrollIntoView({behavior: 'smooth', block: 'center', inline: 'nearest'})
	el.classList.remove(HIGHLIGHT_CLASS)
	// Re-apply on next frame so the animation restarts even if already running.
	requestAnimationFrame(() => {
		el.classList.add(HIGHLIGHT_CLASS)
		window.setTimeout(() => el.classList.remove(HIGHLIGHT_CLASS), HIGHLIGHT_DURATION_MS)
	})
}
