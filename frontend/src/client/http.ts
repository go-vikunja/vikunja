import {client} from '@/client/generated/client.gen'
import type {ResolvedRequestOptions} from '@/client/generated/client/types.gen'
import {
	assertClientRequestContext,
	assertClientRequestMatchesContext,
	captureClientRequestContext,
	type ClientRequestContext,
} from '@/client/requestContext'
import {getToken, getTokenIdentity, refreshToken} from '@/helpers/auth'
import {getApiV2BaseUrl} from '@/helpers/fetcher'
import {AUTH_TYPES} from '@/modelTypes/IUser'

async function getProblemCode(response: Response): Promise<number | null> {
	try {
		const problem = await response.clone().json() as {code?: unknown}
		return typeof problem.code === 'number' ? problem.code : null
	} catch {
		return null
	}
}

const requestContexts = new WeakMap<Request, ClientRequestContext>()
const retryRequests = new WeakMap<Request, {
	request: Request
	token: string
	identity: {id: number; type: number} | null
}>()

function isStreamingResponse(response: Response, options: ResolvedRequestOptions): boolean {
	return response.ok && (options.parseAs === 'stream' ||
		(options.parseAs === 'auto' && response.headers.get('Content-Type') === null)
	)
}

async function fenceResponseBody(
	response: Response,
	context: ClientRequestContext | undefined,
	options: ResolvedRequestOptions,
): Promise<Response> {
	if (!context) {
		return response
	}

	assertClientRequestContext(context)
	if (response.body === null || isStreamingResponse(response, options)) {
		return response
	}

	await response.clone().arrayBuffer()
	assertClientRequestContext(context)
	return response
}

export function configureApiClient(): void {
	client.setConfig({
		baseUrl: getApiV2BaseUrl(),
		credentials: 'include',
		throwOnError: true,
	})
	client.interceptors.request.clear()
	client.interceptors.response.clear()
	client.interceptors.error.clear()

	client.interceptors.request.use((request) => {
		const context = captureClientRequestContext()
		assertClientRequestMatchesContext(request, context)

		let managedRequest = request
		if (request.headers.has('Authorization')) {
			requestContexts.set(managedRequest, context)
			return managedRequest
		}

		const token = getToken()
		if (token) {
			const headers = new Headers(request.headers)
			headers.set('Authorization', `Bearer ${token}`)
			managedRequest = new Request(request, {headers})
			retryRequests.set(managedRequest, {
				request: managedRequest.clone(),
				token,
				identity: getTokenIdentity(token),
			})
		}
		requestContexts.set(managedRequest, context)
		return managedRequest
	})

	client.interceptors.response.use(async (response, request, options: ResolvedRequestOptions) => {
		const context = requestContexts.get(request)
		requestContexts.delete(request)
		response = await fenceResponseBody(response, context, options)

		const retryRequest = retryRequests.get(request)
		retryRequests.delete(request)
		if (!retryRequest || response.status !== 401) {
			return response
		}
		const problemCode = await getProblemCode(response)
		if (context) {
			assertClientRequestContext(context)
		}
		if (problemCode !== 11) {
			return response
		}

		const originalIdentity = retryRequest.identity
		if (originalIdentity?.type !== AUTH_TYPES.USER) {
			return response
		}

		const hasOriginalIdentity = (token: string | null) => {
			const identity = getTokenIdentity(token)
			return identity?.id === originalIdentity.id && identity.type === originalIdentity.type
		}

		let replacementToken = getToken()
		if (!replacementToken || !hasOriginalIdentity(replacementToken)) {
			return response
		}

		if (replacementToken === retryRequest.token) {
			try {
				await refreshToken(true)
			} catch {
				if (context) {
					assertClientRequestContext(context)
				}
				return response
			}
			if (context) {
				assertClientRequestContext(context)
			}

			replacementToken = getToken()
			if (!replacementToken || !hasOriginalIdentity(replacementToken)) {
				return response
			}
		}

		const headers = new Headers(retryRequest.request.headers)
		headers.set('Authorization', `Bearer ${replacementToken}`)
		const retry = new Request(retryRequest.request, {headers})
		const retryResponse = await (options.fetch ?? globalThis.fetch)(retry)
		return fenceResponseBody(retryResponse, context, options)
	})
}
