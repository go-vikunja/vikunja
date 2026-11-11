import DOMPurify from 'dompurify'

// Forced rather than allowed: target leaks window.opener, and a description's own rel could drop noopener.
DOMPurify.addHook('afterSanitizeAttributes', node => {
	if (node.hasAttribute('target')) {
		node.setAttribute('rel', 'noopener noreferrer')
	}
})

export function sanitizeProjectDescription(description: string): string {
	return DOMPurify.sanitize(description, {
		ADD_ATTR: ['target'],
		FORBID_ATTR: ['style'],
		FORBID_TAGS: ['style', 'form', 'input', 'button', 'select', 'textarea'],
	})
}
