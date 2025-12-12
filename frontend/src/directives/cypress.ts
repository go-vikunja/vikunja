import type {Directive} from 'vue'

declare global {
	interface Window {
		Cypress: object;
		TESTING?: boolean;
	}
}

// Check if testing is enabled at runtime
// In dev mode, always enable. In production, check window.TESTING which can be
// injected into index.html before serving (e.g., by CI or test runner)
function isTestingEnabled(): boolean {
	return import.meta.env.DEV || window.TESTING === true
}

const cypressDirective = <Directive<HTMLElement,string>>{
	mounted(el, {arg, value}) {
		if (!isTestingEnabled()) {
			return
		}
		const testingId = arg || value
		if (testingId) {
			el.setAttribute('data-cy', testingId)
		}
	},
	beforeUnmount(el) {
		if (isTestingEnabled()) {
			el.removeAttribute('data-cy')
		}
	},
}

export default cypressDirective
