import type {Directive} from 'vue'
import {install, uninstall} from '@/helpers/shortcut'

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
