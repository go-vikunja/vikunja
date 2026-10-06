import type {VikunjaErrorModel} from '@/client/generated'

// A thrown plain object reaches Sentry without a title or stack, so every failed response becomes one of these.
export class ApiError extends Error implements VikunjaErrorModel {
	declare readonly status: number
	declare readonly detail: string
	declare readonly code?: number
	declare readonly title?: string
	declare readonly errors?: VikunjaErrorModel['errors']
	declare readonly i18n_params?: VikunjaErrorModel['i18n_params']
	declare readonly type?: string
	declare readonly instance?: string

	constructor(problem: VikunjaErrorModel & {status: number, detail: string}) {
		super(problem.detail)
		Object.assign(this, problem)
		this.name = 'ApiError'
	}
}

function isWireProblem(error: unknown): error is Record<string, unknown> {
	return typeof error === 'object' && error !== null && Object.getPrototypeOf(error) === Object.prototype
}

function nonEmptyString(value: unknown): string | undefined {
	return typeof value === 'string' && value.trim() !== '' ? value.trim() : undefined
}

// Echo rejects some /api/v2 requests before Huma (rate limits, invalid JWT) with a v1-shaped body
// carrying `message` but no status or detail. Proxies and Go's default mux answer with plain text
// ("Bad Gateway", "404 page not found"), which the generated client throws as a bare string.
export function normalizeProblemBody(error: unknown, response: Response | undefined): unknown {
	if (!response || response.ok || error instanceof Error) {
		return error
	}
	const body: Record<string, unknown> = isWireProblem(error) ? error : {}
	return new ApiError({
		...body as VikunjaErrorModel,
		status: response.status,
		detail: nonEmptyString(body.detail)
			?? nonEmptyString(body.message)
			?? nonEmptyString(error)
			?? (response.statusText || `HTTP ${response.status}`),
	})
}
