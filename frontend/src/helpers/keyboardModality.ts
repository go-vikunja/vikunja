// :focus-visible matches any focus on text inputs (incl. mouse clicks), so keyboard
// modality is tracked separately via a class and gated on in CSS instead.
export function setupKeyboardModality() {
	window.addEventListener('keydown', (e) => {
		if (e.key === 'Tab') {
			document.documentElement.classList.add('user-is-tabbing')
		}
	})

	window.addEventListener('pointerdown', () => {
		document.documentElement.classList.remove('user-is-tabbing')
	}, {passive: true})
}
