import type {Directive} from 'vue'
import {install, uninstall} from '@github/hotkey'
import {isAppleDevice} from '@/helpers/isAppleDevice'

const directive = <Directive<HTMLElement,string>>{
	mounted(el, {value}) {
		if(value === '') {
			return
		}
		if (isAppleDevice() && value.includes('Control')) {
			value = value.replace('Control', 'Meta')
		}
		install(el, value)
	},
	beforeUnmount(el) {
		uninstall(el)
	},
}

export default directive
