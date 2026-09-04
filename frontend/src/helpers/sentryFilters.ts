import {AxiosError} from 'axios'

// Failed requests are surfaced to the user through the UI already, and an
// expired session (401 on token refresh) is expected rather than a bug.
// Errors wrapping one of them via `cause` count too.
const MAX_CAUSE_DEPTH = 10

// Thrown when a user has an old index.html open and a deploy changed the asset
// hashes. handleChunkLoadErrors() reloads the page instead.
const CHUNK_LOAD_ERROR_PATTERNS = [
	/failed to fetch dynamically imported module/i,
	/error loading dynamically imported module/i,
	/importing a module script failed/i,
	/unable to preload css/i,
]

type SentryEventLike = {
	message?: string
	exception?: {
		values?: {value?: string}[]
	}
}

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

export function isChunkLoadError(message: unknown): boolean {
	return typeof message === 'string'
		&& CHUNK_LOAD_ERROR_PATTERNS.some(pattern => pattern.test(message))
}

export function shouldDropEvent(originalException: unknown, event?: SentryEventLike): boolean {
	if (isChunkLoadError(event?.message)) {
		return true
	}

	if (event?.exception?.values?.some(value => isChunkLoadError(value?.value))) {
		return true
	}

	let current = originalException

	for (let depth = 0; depth < MAX_CAUSE_DEPTH && current; depth++) {
		if (isRequestError(current) || isChunkLoadError((current as {message?: unknown}).message)) {
			return true
		}

		current = (current as {cause?: unknown}).cause
	}

	return false
}
