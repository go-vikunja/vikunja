import {createRandomID} from '@/helpers/randomId'
import tippy from 'tippy.js'
import {nextTick} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'

export default function inputPrompt(pos: ClientRect, oldValue: string = ''): Promise<string> {
	return new Promise((resolve) => {
		const id = 'link-input-' + createRandomID()

		const linkPopup = tippy('body', {
			getReferenceClientRect: () => pos,
			appendTo: () => document.body,
			content: `<div><input class="input" placeholder="URL" id="${id}" value="${oldValue}"/></div>`,
			showOnCreate: true,
			interactive: true,
			trigger: 'manual',
			placement: 'top-start',
			allowHTML: true,
		})

		linkPopup[0].show()

		nextTick(() => document.getElementById(id)?.focus())

		document.getElementById(id)?.addEventListener('keydown', event => {
			const hotkeyString = eventToHotkeyString(event)
			if (hotkeyString !== 'Enter') {
				return
			}

			const url = (event.target as HTMLInputElement)?.value || ''

			resolve(url)

			linkPopup[0].hide()
		})

	})
}
