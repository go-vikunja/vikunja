import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'

const STORAGE_KEY_PREFIX = 'editorDraft'

/**
 * Save editor content to local storage
 */
export function saveEditorDraft(storageKey: string, content: string) {
	if (!storageKey) {
		return
	}

	const key = `${STORAGE_KEY_PREFIX}-${storageKey}`

	try {
		if (!content || isEditorContentEmpty(content)) {
			// Remove empty drafts
			localStorage.removeItem(key)
			return
		}

		localStorage.setItem(key, content)
	} catch (error) {
		console.warn('Failed to save editor draft:', error)
	}
}

/**
 * Load editor content from local storage
 */
export function loadEditorDraft(storageKey: string): string | null {
	if (!storageKey) {
		return null
	}

	const key = `${STORAGE_KEY_PREFIX}-${storageKey}`
	
	try {
		return localStorage.getItem(key)
	} catch (error) {
		console.warn('Failed to load editor draft:', error)
		return null
	}
}

/**
 * Clear editor content from local storage
 */
export function clearEditorDraft(storageKey: string) {
	if (!storageKey) {
		return
	}

	const key = `${STORAGE_KEY_PREFIX}-${storageKey}`
	
	try {
		localStorage.removeItem(key)
	} catch (error) {
		console.warn('Failed to clear editor draft:', error)
	}
}
