import {afterEach, describe, expect, it, vi} from 'vitest'

import {publicClient} from './publicClient'

describe('publicClient', () => {
	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it('stamps the response status onto an Echo error body', async () => {
		vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({message: 'Too Many Requests'}), {
			status: 429,
			headers: {'Content-Type': 'application/json'},
		})))

		await expect(publicClient.get({url: 'https://api.example.com/api/v2/token'})).rejects.toEqual({
			message: 'Too Many Requests',
			detail: 'Too Many Requests',
			status: 429,
		})
	})

	it('sends the session cookie without an Authorization header', async () => {
		const requests: Request[] = []
		vi.stubGlobal('fetch', vi.fn(async (request: Request) => {
			requests.push(request)
			return new Response(JSON.stringify({ok: true}), {
				status: 200,
				headers: {'Content-Type': 'application/json'},
			})
		}))

		await publicClient.get({url: 'https://api.example.com/api/v2/token'})

		expect(requests[0].credentials).toBe('include')
		expect(requests[0].headers.has('Authorization')).toBe(false)
	})
})
