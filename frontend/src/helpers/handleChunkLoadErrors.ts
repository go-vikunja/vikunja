import {isChunkLoadError} from './sentryFilters'

const LAST_RELOAD_KEY = 'chunkLoadErrorReloadedAt'
const RELOAD_COOLDOWN = 60 * 1000

function readLastReload(): number {
	try {
		return Number(sessionStorage.getItem(LAST_RELOAD_KEY))
	} catch {
		return 0
	}
}

/**
 * Reloading only helps when the page is stale, so a second failure right after
 * a reload means something else is broken — let that one through to Sentry
 * instead of reloading forever.
 */
export function canReloadForChunkLoadError(now: number = Date.now()): boolean {
	const lastReload = readLastReload()

	if (!Number.isFinite(lastReload) || lastReload <= 0) {
		return true
	}

	return now - lastReload >= RELOAD_COOLDOWN
}

export function markChunkLoadErrorReload(now: number = Date.now()) {
	try {
		sessionStorage.setItem(LAST_RELOAD_KEY, String(now))
	} catch {
		// A blocked sessionStorage only costs us the loop guard.
	}
}

function reloadForStaleChunk(event: Event) {
	if (!canReloadForChunkLoadError()) {
		return
	}

	event.preventDefault()
	markChunkLoadErrorReload()
	window.location.reload()
}

export function handleChunkLoadErrors() {
	window.addEventListener('vite:preloadError', reloadForStaleChunk)

	// Dynamic imports that don't go through vite's preload helper — and the MIME
	// type errors browsers raise when index.html is served for a gone chunk —
	// only ever surface as a rejected promise.
	window.addEventListener('unhandledrejection', event => {
		if (!isChunkLoadError((event.reason as {message?: unknown} | undefined)?.message)) {
			return
		}

		reloadForStaleChunk(event)
	})
}
