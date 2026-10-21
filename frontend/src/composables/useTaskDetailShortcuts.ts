import {ref, onMounted, onBeforeUnmount} from 'vue'
import {eventToShortcutString, isFormField} from '@/helpers/shortcut'
import {SHORTCUTS} from '@/constants/shortcuts'

import {getTaskIdentifier} from '@/helpers/task'
import type {Task as ITask} from '@/client/generated'

interface UseTaskDetailShortcutsOptions {
	task: () => ITask
	taskTitle: () => string
	onSave: () => void
	onToggleDone: () => void
}

async function copySavely(value: string) {

	try {
		await navigator.clipboard.writeText(value)
	} catch(e) {
		console.error('could not write to clipboard', e)
	}
}

export function useTaskDetailShortcuts({
	task,
	taskTitle,
	onSave,
	onToggleDone,
}: UseTaskDetailShortcutsOptions) {
	const dotKeyPressedTimes = ref(0)
	const dotKeyCopyValue = ref('')
	let dotKeyPressedTimeout: ReturnType<typeof setTimeout> | null = null

	function resetDotKeyPressed() {
		dotKeyPressedTimes.value = 0
		dotKeyCopyValue.value = ''
		if (dotKeyPressedTimeout !== null) {
			clearTimeout(dotKeyPressedTimeout)
			dotKeyPressedTimeout = null
		}
	}

	// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
	async function handleTaskHotkey(event: KeyboardEvent) {
		const shortcutString = eventToShortcutString(event)
		if (!shortcutString) return

		if (shortcutString === 'Control+KeyS' || shortcutString === 'Meta+KeyS') {
			event.preventDefault()
			onSave()
			return
		}

		if (shortcutString === SHORTCUTS.taskDetail.done) {
			if (event.defaultPrevented || event.isComposing || event.repeat || isFormField(event.target)) return
			event.preventDefault()
			onToggleDone()
			return
		}

		const target = event.target as HTMLElement

		if (
			target.tagName.toLowerCase() === 'input' ||
			target.tagName.toLowerCase() === 'textarea' ||
			target.contentEditable === 'true'
		) {
			return
		}

		if (shortcutString === 'Control+Period') {
			await copySavely(window.location.href)
			return
		}

		if (shortcutString === 'Period') {
			dotKeyPressedTimes.value++
			if (dotKeyPressedTimeout !== null) {
				clearTimeout(dotKeyPressedTimeout)
			}
			dotKeyPressedTimeout = setTimeout(async () => {
				await copySavely(dotKeyCopyValue.value)
				resetDotKeyPressed()
			}, 300)

			switch (dotKeyPressedTimes.value) {
				case 1:
					dotKeyCopyValue.value = getTaskIdentifier(task())
					break
				case 2:
					dotKeyCopyValue.value += ' - ' + taskTitle()
					break
				case 3:
					dotKeyCopyValue.value += ' - ' + window.location.href
					break
				default:
					resetDotKeyPressed()
			}
		}
	}

	onMounted(() => {
		document.addEventListener('keydown', handleTaskHotkey)
	})

	onBeforeUnmount(() => {
		document.removeEventListener('keydown', handleTaskHotkey)
		resetDotKeyPressed()
	})
}
