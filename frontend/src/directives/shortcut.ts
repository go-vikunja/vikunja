import type {Directive} from 'vue'
import {install, uninstall} from '@github/hotkey'

const directive = <Directive<HTMLElement, string | string[]>>{
	mounted(el, { value }) {
		if (typeof value === 'string')
		{
			if (value !== '') install(el, value)
			return
		}

		if (value.length === 0) return

		install(el, value.join(', '))
	},
	beforeUnmount(el) {
		uninstall(el)
	},
}

export default directive
