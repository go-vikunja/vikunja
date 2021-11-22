import {Directive} from 'vue'

declare global {
	interface Window {
		Cypress: object;
	}
}

const cypressDirective: Directive = {
	mounted(el, {value}) {
		if (
			(window.Cypress || import.meta.env.DEV) &&
			value
		) {
			el.setAttribute('data-cy', value)
		}
	},
	beforeUnmount(el) {
		el.removeAttribute('data-cy')
	},
}

export default cypressDirective
