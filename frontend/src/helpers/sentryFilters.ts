import {AxiosError} from 'axios'

// Failed requests are surfaced to the user through the UI already, and an
// expired session (401 on token refresh) is expected rather than a bug.
// Errors wrapping one of them via `cause` count too.
const MAX_CAUSE_DEPTH = 10

function isRequestError(e: unknown): boolean {
	if (e instanceof AxiosError) {
		return true
	}

	if (typeof e !== 'object' || e === null) {
		return false
	}

	return typeof (e as {code?: unknown}).code !== 'undefined'
		&& typeof (e as {message?: unknown}).message !== 'undefined'
}

export function shouldDropEvent(originalException: unknown): boolean {
	let current = originalException

	for (let depth = 0; depth < MAX_CAUSE_DEPTH && current; depth++) {
		if (isRequestError(current)) {
			return true
		}

		current = (current as {cause?: unknown}).cause
	}

	return false
}
