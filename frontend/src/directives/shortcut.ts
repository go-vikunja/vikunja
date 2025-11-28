import type {Directive} from 'vue'
import {install, uninstall} from '@github/hotkey'
import {useShortcutManager} from '@/composables/useShortcutManager'

const directive = <Directive<HTMLElement,string>>{
	mounted(el, {value}) {
		if(value === '') {
			return
		}

		// Support both old format (direct keys) and new format (actionId)
		const shortcutManager = useShortcutManager()
		const hotkeyString = value.startsWith('.')
			? shortcutManager.getHotkeyString(value.slice(1))  // New format: actionId (remove leading dot)
			: value                                             // Old format: direct keys (backwards compat)

		if (!hotkeyString) return

		install(el, hotkeyString)

		// Store for cleanup and updates
		el.dataset.shortcutActionId = value
	},
	updated(el, {value, oldValue}) {
		if (value === oldValue) return

		// Reinstall with new shortcut
		uninstall(el)

		if(value === '') {
			return
		}

		const shortcutManager = useShortcutManager()
		const hotkeyString = value.startsWith('.')
			? shortcutManager.getHotkeyString(value.slice(1))
			: value

		if (!hotkeyString) return
		install(el, hotkeyString)
		el.dataset.shortcutActionId = value
	},
	beforeUnmount(el) {
		uninstall(el)
	},
}

export default directive
