import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {client} from './generated/client.gen'
import {configureApiClient} from './http'

const auth = vi.hoisted(() => ({
	token: null as string | null,
	type: null as number | null,
	id: 1,
	identities: new Map<string, {id: number; type: number}>(),
	refreshToken: vi.fn(),
}))

vi.mock('@/helpers/auth', () => ({
	getToken: () => auth.token,
	getTokenType: () => auth.type,
	getTokenIdentity: (token: string | null) => token
		? auth.identities.get(token) ?? (auth.type === null ? null : {id: auth.id, type: auth.type})
		: null,
	refreshToken: auth.refreshToken,
}))

const problem = (code: number, detail = 'request failed') => new Response(JSON.stringify({code, detail}), {
	status: 401,
	headers: {'Content-Type': 'application/problem+json'},
})

const ok = () => new Response(JSON.stringify({ok: true}), {
	status: 200,
	headers: {'Content-Type': 'application/json'},
})

const BrowserRequest = globalThis.Request

class BodyAwareRequest extends BrowserRequest {
	constructor(input: RequestInfo | URL, init?: RequestInit) {
		if (input instanceof BrowserRequest && input.body !== null && input.bodyUsed) {
			throw new TypeError('Body has already been consumed')
		}
		super(input, init)
	}
}

function deferred<T>() {
	let resolve!: (value: T) => void
	const promise = new Promise<T>((resolvePromise) => {
		resolve = resolvePromise
	})
	return {promise, resolve}
}

describe('configureApiClient', () => {
	let requests: Request[]
	let responses: Response[]

	beforeEach(() => {
		window.API_URL = 'https://api.example.com/root/api/v1'
		auth.token = null
		auth.type = null
		auth.id = 1
		auth.identities.clear()
		auth.refreshToken.mockReset()
		requests = []
		responses = [ok()]
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			return responses.shift() ?? ok()
		}))
		configureApiClient()
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it('configures the v2 base URL and includes credentials', async () => {
		await client.get({url: '/probe'})

		expect(requests[0].url).toBe('https://api.example.com/root/api/v2/probe')
		expect(requests[0].credentials).toBe('include')
	})

	it('adds the current bearer token', async () => {
		auth.token = 'current-token'

		await client.get({url: '/probe'})

		expect(requests[0].headers.get('Authorization')).toBe('Bearer current-token')
	})

	it('preserves explicit basic authorization', async () => {
		auth.token = 'session-token'

		await client.get({
			url: '/probe',
			headers: {Authorization: 'Basic client-credentials'},
		})

		expect(requests[0].headers.get('Authorization')).toBe('Basic client-credentials')
	})

	it('does not refresh explicit bearer authorization', async () => {
		auth.token = 'session-token'
		auth.type = 1
		responses = [problem(11)]

		await expect(client.get({
			url: '/probe',
			headers: {Authorization: 'Bearer api-token'},
		})).rejects.toMatchObject({code: 11})

		expect(requests[0].headers.get('Authorization')).toBe('Bearer api-token')
		expect(auth.refreshToken).not.toHaveBeenCalled()
		expect(requests).toHaveLength(1)
	})

	it('refreshes an expired user token and retries once', async () => {
		auth.token = 'expired-token'
		auth.type = 1
		responses = [problem(11), ok()]
		auth.refreshToken.mockImplementation(async () => {
			auth.token = 'replacement-token'
		})

		await client.get({url: '/probe'})

		expect(auth.refreshToken).toHaveBeenCalledOnce()
		expect(auth.refreshToken).toHaveBeenCalledWith(true)
		expect(requests).toHaveLength(2)
		expect(requests[1].headers.get('Authorization')).toBe('Bearer replacement-token')
	})

	it.each(['POST', 'PUT', 'PATCH'] as const)('replays a consumed %s body after refreshing', async (method) => {
		auth.token = 'expired-token'
		auth.type = 1
		responses = [problem(11), ok()]
		auth.refreshToken.mockImplementation(async () => {
			auth.token = 'replacement-token'
		})
		const bodies: string[] = []
		vi.stubGlobal('Request', BodyAwareRequest)
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			bodies.push(await request.text())
			return responses.shift() ?? ok()
		}))

		await client.request({
			method,
			url: '/probe',
			body: {message: 'keep me'},
		})

		expect(requests).toHaveLength(2)
		expect(bodies).toEqual([
			JSON.stringify({message: 'keep me'}),
			JSON.stringify({message: 'keep me'}),
		])
		expect(auth.refreshToken).toHaveBeenCalledOnce()
	})

	it('reuses a refreshed token for a delayed 401 response', async () => {
		auth.token = 'expired-token'
		auth.type = 1
		auth.refreshToken.mockImplementation(async () => {
			auth.token = 'replacement-token'
		})
		const firstResponse = deferred<Response>()
		const secondResponse = deferred<Response>()
		let expiredRequests = 0
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			if (request.headers.get('Authorization') !== 'Bearer expired-token') {
				return ok()
			}
			expiredRequests++
			return expiredRequests === 1 ? firstResponse.promise : secondResponse.promise
		}))

		const first = client.get({url: '/first'})
		const second = client.get({url: '/second'})
		await vi.waitFor(() => expect(requests).toHaveLength(2))

		firstResponse.resolve(problem(11))
		await first
		expect(auth.refreshToken).toHaveBeenCalledOnce()

		secondResponse.resolve(problem(11))
		await second

		expect(auth.refreshToken).toHaveBeenCalledOnce()
		expect(requests).toHaveLength(4)
		expect(requests[3].headers.get('Authorization')).toBe('Bearer replacement-token')
	})

	it('does not replay a delayed mutation after the user changes', async () => {
		auth.token = 'user-a-token'
		auth.type = 1
		auth.identities.set('user-a-token', {id: 1, type: 1})
		auth.identities.set('user-b-token', {id: 2, type: 1})
		const firstResponse = deferred<Response>()
		const bodies: string[] = []
		vi.stubGlobal('Request', BodyAwareRequest)
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			bodies.push(await request.text())
			return firstResponse.promise
		}))

		const mutation = client.request({
			method: 'POST',
			url: '/probe',
			body: {message: 'user a data'},
		})
		await vi.waitFor(() => expect(requests).toHaveLength(1))

		auth.token = 'user-b-token'
		firstResponse.resolve(problem(11))

		await expect(mutation).rejects.toMatchObject({code: 11})
		expect(auth.refreshToken).not.toHaveBeenCalled()
		expect(requests).toHaveLength(1)
		expect(bodies).toEqual([JSON.stringify({message: 'user a data'})])
	})

	it('does not replay a mutation when the user changes during refresh', async () => {
		auth.token = 'user-a-token'
		auth.type = 1
		auth.identities.set('user-a-token', {id: 1, type: 1})
		auth.identities.set('user-b-token', {id: 2, type: 1})
		const refresh = deferred<void>()
		auth.refreshToken.mockImplementation(() => refresh.promise)
		const bodies: string[] = []
		vi.stubGlobal('Request', BodyAwareRequest)
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			bodies.push(await request.text())
			return problem(11)
		}))

		const mutation = client.request({
			method: 'POST',
			url: '/probe',
			body: {message: 'user a data'},
		})
		await vi.waitFor(() => expect(auth.refreshToken).toHaveBeenCalledOnce())

		auth.token = 'user-b-token'
		refresh.resolve()

		await expect(mutation).rejects.toMatchObject({code: 11})
		expect(requests).toHaveLength(1)
		expect(bodies).toEqual([JSON.stringify({message: 'user a data'})])
	})

	it('does not refresh link-share tokens', async () => {
		auth.token = 'link-share-token'
		auth.type = 2
		responses = [problem(11)]

		await expect(client.get({url: '/probe'})).rejects.toMatchObject({code: 11})

		expect(auth.refreshToken).not.toHaveBeenCalled()
		expect(requests).toHaveLength(1)
	})

	it('does not refresh without a token', async () => {
		auth.type = 1
		responses = [problem(11)]

		await expect(client.get({url: '/probe'})).rejects.toMatchObject({code: 11})

		expect(auth.refreshToken).not.toHaveBeenCalled()
		expect(requests).toHaveLength(1)
	})

	it('does not retry other 401 responses', async () => {
		auth.token = 'current-token'
		auth.type = 1
		responses = [problem(1017)]

		await expect(client.get({url: '/probe'})).rejects.toMatchObject({code: 1017})

		expect(auth.refreshToken).not.toHaveBeenCalled()
		expect(requests).toHaveLength(1)
	})

	it('returns the second 401 problem after one retry', async () => {
		auth.token = 'expired-token'
		auth.type = 1
		responses = [problem(11, 'expired'), problem(11, 'still expired')]
		auth.refreshToken.mockImplementation(async () => {
			auth.token = 'replacement-token'
		})

		await expect(client.get({url: '/probe'})).rejects.toMatchObject({
			code: 11,
			detail: 'still expired',
		})

		expect(auth.refreshToken).toHaveBeenCalledOnce()
		expect(requests).toHaveLength(2)
	})
})
