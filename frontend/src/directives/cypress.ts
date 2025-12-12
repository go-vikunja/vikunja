import type {Directive} from 'vue'

declare global {
	interface Window {
		Cypress: object;
	}
}

const cypressDirective = <Directive<HTMLElement,string>>{
	mounted(el, {arg, value}) {
		const testingId = arg || value
		// Always add data-cy attributes - they're harmless metadata and ensure
		// tests work in both dev mode and production builds (e.g., Playwright in CI)
		if (testingId) {
			el.setAttribute('data-cy', testingId)
		}
	},
	beforeUnmount(el) {
		el.removeAttribute('data-cy')
	},
}

export default cypressDirective
