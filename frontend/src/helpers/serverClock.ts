import {MILLISECONDS_A_MINUTE, MILLISECONDS_A_SECOND} from '@/constants/date'

const OFFSET_STORAGE_KEY = 'serverClockOffset'

// Server minus browser time, measured when the server hands out a token.
let offsetMs: number | null = null
let skewWarned = false

// Only call with a token the server just issued; tokens without iat reset the offset to 0.
export function recordServerClock(issuedAtSeconds: unknown, persist: boolean) {
	offsetMs = typeof issuedAtSeconds === 'number'
		? issuedAtSeconds * MILLISECONDS_A_SECOND - Date.now()
		: 0
	if (persist) {
		localStorage.setItem(OFFSET_STORAGE_KEY, String(offsetMs))
	}

	if (!skewWarned && Math.abs(offsetMs) > MILLISECONDS_A_MINUTE) {
		skewWarned = true
		const seconds = Math.round(Math.abs(offsetMs) / MILLISECONDS_A_SECOND)
		const direction = offsetMs < 0 ? 'ahead of' : 'behind'
		console.warn(
			`Clock skew detected: the browser clock is ${seconds} s ${direction} the server clock. ` +
			'Token expiry is checked against server time.',
		)
	}
}

export function clearServerClock() {
	offsetMs = null
	localStorage.removeItem(OFFSET_STORAGE_KEY)
}

export function serverNowSeconds(): number {
	if (offsetMs === null) {
		offsetMs = Number(localStorage.getItem(OFFSET_STORAGE_KEY)) || 0
	}
	return (Date.now() + offsetMs) / MILLISECONDS_A_SECOND
}
