import {ref, onMounted, onBeforeUnmount} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'

import {getTaskIdentifier} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

interface UseTaskDetailShortcutsOptions {
	task: () => ITask
	taskTitle: () => string
	onSave: () => void
}

export function useTaskDetailShortcuts({
	task,
	taskTitle,
	onSave,
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
	function handleTaskHotkey(event: KeyboardEvent) {
		const hotkeyString = eventToHotkeyString(event)
		if (!hotkeyString) return

		if (hotkeyString === 'Control+s' || hotkeyString === 'Meta+s') {
			event.preventDefault()
			onSave()
		}

		if (hotkeyString === '.') {
			dotKeyPressedTimes.value++
			if (dotKeyPressedTimeout !== null) {
				clearTimeout(dotKeyPressedTimeout)
			}
			dotKeyPressedTimeout = setTimeout(() => {
				navigator.clipboard.writeText(dotKeyCopyValue.value)
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
