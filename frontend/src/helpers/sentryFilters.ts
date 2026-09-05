import {AxiosError} from 'axios'

// Failed requests are surfaced to the user through the UI already, and an
// expired session (401 on token refresh) is expected rather than a bug.
// Errors wrapping one of them via `cause` count too.
const MAX_CAUSE_DEPTH = 10

// Thrown when a user has an old index.html open and a deploy changed the asset
// hashes. handleChunkLoadErrors() reloads the page instead.
// The MIME variants are what browsers report when the server answers the request
// for a gone chunk with index.html instead of a 404.
const CHUNK_LOAD_ERROR_PATTERNS = [
	/failed to fetch dynamically imported module/i,
	/error loading dynamically imported module/i,
	/importing a module script failed/i,
	/unable to preload css/i,
	/is not a valid javascript mime type/i,
	/was blocked because of a disallowed mime type/i,
	/expected a javascript(-or-wasm)? module script but the server responded with a mime type/i,
]

// Injected into our page by browser extensions and by in-app webviews
// (Instagram, WeChat, ...). Their scripts run on our origin but none of it is
// our code, and we can't fix any of it.
const THIRD_PARTY_INJECTION_PATTERNS = [
	/runtime\.sendmessage\(\)/i,
	/window\.webkit\.messagehandlers/i,
	/error invoking postmessage/i,
	/weixinpostmessagehandlers/i,
]

const THIRD_PARTY_URL_PATTERN = /^(?:(?:chrome|moz|safari-web|safari|ms-browser)-extension:|iabjs:)/i

type SentryEventLike = {
	message?: string
	exception?: {
		values?: {
			type?: string
			value?: string
			stacktrace?: {
				frames?: {filename?: string}[]
			}
		}[]
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

function isNoisyMessage(message: unknown): boolean {
	return isChunkLoadError(message)
		|| (typeof message === 'string' && THIRD_PARTY_INJECTION_PATTERNS.some(pattern => pattern.test(message)))
}

function hasThirdPartyTopFrame(event?: SentryEventLike): boolean {
	return event?.exception?.values?.some(value => {
		const frames = value?.stacktrace?.frames
		// Sentry orders frames oldest first, so the one that threw is last.
		const filename = frames?.[frames.length - 1]?.filename

		return typeof filename === 'string' && THIRD_PARTY_URL_PATTERN.test(filename)
	}) ?? false
}

// Sentry sends an event even when it has neither a message nor an exception to
// put in it, e.g. after a `Promise.reject()` with no value. There is nothing to
// act on in those.
function hasNothingToReport(event?: SentryEventLike): boolean {
	if (!event || event.message) {
		return false
	}

	const values = event.exception?.values

	return !values?.length || values.every(value => !value?.value && !value?.type)
}

export function shouldDropEvent(originalException: unknown, event?: SentryEventLike): boolean {
	if (isNoisyMessage(event?.message)) {
		return true
	}

	if (event?.exception?.values?.some(value => isNoisyMessage(value?.value))) {
		return true
	}

	if (hasThirdPartyTopFrame(event)) {
		return true
	}

	if (hasNothingToReport(event)) {
		return true
	}

	let current = originalException

	for (let depth = 0; depth < MAX_CAUSE_DEPTH && current; depth++) {
		if (isRequestError(current) || isNoisyMessage((current as {message?: unknown}).message)) {
			return true
		}

		current = (current as {cause?: unknown}).cause
	}

	return false
}
