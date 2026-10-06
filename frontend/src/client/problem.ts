// Only a parsed JSON body; Error instances come from the transport or our own interceptors.
function isWireProblem(error: unknown): error is Record<string, unknown> {
	return typeof error === 'object' && error !== null && Object.getPrototypeOf(error) === Object.prototype
}

// Echo rejects some /api/v2 requests before Huma (rate limits, invalid JWT) and answers with a
// v1-shaped body, so the top-level status/detail of api.md only holds once we stamp it on.
// Proxies and Go's default mux answer with plain text ("Bad Gateway", "404 page not found"), which
// the generated client throws as a bare string.
export function normalizeProblemBody(error: unknown, response: Response | undefined): unknown {
	if (!response || response.ok) {
		return error
	}
	if (typeof error === 'string') {
		return {
			status: response.status,
			detail: error.trim() || response.statusText || `HTTP ${response.status}`,
		}
	}
	if (!isWireProblem(error) || typeof error.status === 'number') {
		return error
	}
	return {
		...error,
		status: response.status,
		detail: error.detail ?? error.message,
	}
}
