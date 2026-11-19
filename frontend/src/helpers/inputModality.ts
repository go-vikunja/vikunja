// Whether the latest input anywhere in the page was a pointer press rather than a key. One pair of
// capture listeners serves every caller, so a page full of popups adds no listeners of its own.
let lastInputWasPointer = false
let listening = false

function listen() {
	if (listening || typeof window === 'undefined') {
		return
	}
	listening = true
	window.addEventListener('pointerdown', () => {
		lastInputWasPointer = true
	}, true)
	window.addEventListener('keydown', () => {
		lastInputWasPointer = false
	}, true)
}

export function useInputModality() {
	listen()
	return {
		lastInputWasPointer: () => lastInputWasPointer,
	}
}
