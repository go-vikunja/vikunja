import type {Directive} from 'vue'
import {install, uninstall} from '@github/hotkey'

const directive = <Directive<HTMLElement,string>>{
	mounted(el, {value}) {
		if(value === '') {
			return
		}
		install(el, value)
	},
	beforeUnmount(el) {
		uninstall(el)
	},
}

export default directive
