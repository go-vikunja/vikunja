import type {Directive} from 'vue'

declare global {
	interface Window {
		Cypress: object;
	}
}

const cypressDirective = <Directive<HTMLElement,string>>{
	mounted(el, {arg, value}) {
		const testingId = arg || value
		if ((window.Cypress || import.meta.env.DEV) && testingId) {
			el.setAttribute('data-cy', testingId)
		}
	},
	beforeUnmount(el) {
		el.removeAttribute('data-cy')
	},
}

export default cypressDirective
