import {ref, onMounted, onBeforeUnmount} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'

import {getTaskIdentifier} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import {comboHotkey} from '@/components/misc/keyboard-shortcuts/shortcuts'

interface UseTaskDetailShortcutsOptions {
	task: () => ITask
	taskTitle: () => string
	onSave: () => void
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
		const hotkeyString = eventToHotkeyString(event)
		if (!hotkeyString) return

		const expectedHotkeys:string[] = comboHotkey('Meta', 's')
		expectedHotkeys.concat(comboHotkey('Control', 's'))
		if (expectedHotkeys.includes(hotkeyString)) {
			event.preventDefault()
			onSave()
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

		if (hotkeyString === 'Control+.') {
			await copySavely(window.location.href)
			return
		}

		if (hotkeyString === '.') {
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
