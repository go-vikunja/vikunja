import {
	afterEach,
	expect,
	it,
	vi,
} from 'vitest'
import {useWebSocket} from './useWebSocket'
import {AUTH_TYPES} from '@/constants/auth'
// expired: null defers to the real expiry check against the token's exp claim.
const auth = vi.hoisted(() => ({tokenType: 1, token: 'token', expired: false as boolean | null}))
const refreshToken = vi.hoisted(() => vi.fn())
const refreshListeners = vi.hoisted(() => [] as (() => void)[])
const RefreshTokenError = vi.hoisted(() => class extends Error {
	constructor(public failure: {kind: string}) { super() }
})
vi.mock('@/helpers/auth', async importOriginal => {
	const actual = await importOriginal<typeof import('@/helpers/auth')>()
	return {
		getToken: () => auth.token,
		getTokenType: () => auth.tokenType,
		isTokenExpired: (token: string | null, marginSeconds?: number) =>
			auth.expired ?? actual.isTokenExpired(token, marginSeconds),
		refreshToken,
		RefreshTokenError,
		onTokenRefreshed: (listener: () => void) => refreshListeners.push(listener),
	}
})
const session = vi.hoisted(() => ({current: true}))
vi.mock('@/client/requestContext', () => ({
	captureClientRequestContext: () => ({}),
	isClientRequestContextCurrent: () => session.current,
}))
class FakeSocket {
	static OPEN = 1
	static CONNECTING = 0
	static instances: FakeSocket[] = []
	readyState = 1
	onopen: (() => void) | null = null
	onmessage: ((event: MessageEvent) => void) | null = null
	onclose: (() => void) | null = null
	onerror: (() => void) | null = null
	send = vi.fn()
	close = vi.fn(() => {
		this.readyState = 3
	})
	constructor(public url: string) { FakeSocket.instances.push(this) }
}

function timerFrame() {
	return new MessageEvent('message', {data: JSON.stringify({
		event: 'timer.created',
		data: {id: 1},
	})})
}

function authSuccessFrame() {
	return new MessageEvent('message', {data: JSON.stringify({
		action: 'auth.success',
		success: true,
	})})
}

function authErrorFrame() {
	return new MessageEvent('message', {data: JSON.stringify({error: 'invalid_token'})})
}

function tokenRefreshed(token: string) {
	auth.token = token
	refreshListeners.forEach(listener => listener())
}

function authenticatedSocket() {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	socket.onmessage?.(authSuccessFrame())
	socket.send.mockClear()
	return {ws, socket}
}

afterEach(() => {
	useWebSocket().disconnect()
	vi.useRealTimers()
	vi.unstubAllGlobals()
	vi.restoreAllMocks()
	FakeSocket.instances = []
	session.current = true
	auth.tokenType = AUTH_TYPES.USER
	auth.token = 'token'
	auth.expired = false
	refreshToken.mockReset()
})

it('refreshes an expired token before authenticating', async () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	auth.expired = true
	refreshToken.mockImplementation(async () => {
		auth.token = 'fresh'
		auth.expired = false
	})
	useWebSocket().connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	await vi.waitFor(() => expect(socket.send).toHaveBeenCalled())

	expect(refreshToken).toHaveBeenCalledTimes(1)
	expect(socket.send).toHaveBeenCalledWith(JSON.stringify({action: 'auth', token: 'fresh'}))
})

it('refreshes a token that expires within a second before authenticating', async () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	auth.expired = null
	auth.token = `header.${btoa(JSON.stringify({exp: Date.now() / 1000 + 1}))}.signature`
	refreshToken.mockImplementation(async () => {
		auth.token = 'fresh'
	})
	useWebSocket().connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	await vi.waitFor(() => expect(socket.send).toHaveBeenCalled())

	expect(refreshToken).toHaveBeenCalledTimes(1)
	expect(socket.send).toHaveBeenCalledWith(JSON.stringify({action: 'auth', token: 'fresh'}))
})

it('reconnects instead of sending the expired token when the refresh fails transiently', async () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	auth.expired = true
	refreshToken.mockRejectedValue(new RefreshTokenError({kind: 'network'}))
	useWebSocket().connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	await vi.waitFor(() => expect(socket.close).toHaveBeenCalled())
	socket.onclose?.()

	expect(socket.send).not.toHaveBeenCalled()
	vi.runOnlyPendingTimers()
	expect(FakeSocket.instances).toHaveLength(2)
})

it('gives up when the refresh is rejected', async () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	auth.expired = true
	refreshToken.mockRejectedValue(new RefreshTokenError({kind: 'rejected'}))
	const ws = useWebSocket()
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	await vi.waitFor(() => expect(socket.send).toHaveBeenCalled())
	socket.onmessage?.(authErrorFrame())

	vi.runOnlyPendingTimers()
	expect(FakeSocket.instances).toHaveLength(1)
	expect(ws.status.value).toBe('idle')
})

it('authenticates with a valid token without refreshing it', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	useWebSocket().connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()

	expect(refreshToken).not.toHaveBeenCalled()
	expect(socket.send).toHaveBeenCalledWith(JSON.stringify({action: 'auth', token: 'token'}))
})

it('routes messages from the current connection to subscribers', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	const onTimer = vi.fn()
	ws.subscribe('timer.created', onTimer)
	socket.onmessage?.(timerFrame())
	expect(onTimer).toHaveBeenCalledTimes(1)
	expect(onTimer).toHaveBeenCalledWith({
		event: 'timer.created',
		data: {id: 1},
	})
})

it('ignores messages and close callbacks from a replaced connection', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const old = FakeSocket.instances[0]
	ws.disconnect()
	ws.connect()
	const current = FakeSocket.instances[1]
	current.onopen?.()
	const onTimer = vi.fn()
	ws.subscribe('timer.created', onTimer)
	old.onmessage?.(timerFrame())
	old.onclose?.()
	expect(onTimer).not.toHaveBeenCalled()
	expect(ws.status.value).toBe('authenticating')
})

it('closes the socket and reconnects after the authenticated session changes', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const stale = FakeSocket.instances[0]
	stale.onopen?.()
	const onTimer = vi.fn()
	ws.subscribe('timer.created', onTimer)
	session.current = false
	stale.onmessage?.(timerFrame())
	expect(onTimer).not.toHaveBeenCalled()
	expect(stale.close).toHaveBeenCalledTimes(1)
	expect(ws.status.value).toBe('idle')
	expect(ws.authenticated.value).toBe(false)

	session.current = true
	vi.runOnlyPendingTimers()
	expect(FakeSocket.instances).toHaveLength(2)
	expect(FakeSocket.instances[1]).not.toBe(stale)
})

it('tears down a socket opened by a previous session instead of reusing it', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const stale = FakeSocket.instances[0]
	stale.onopen?.()
	stale.onmessage?.(authSuccessFrame())
	expect(ws.authenticated.value).toBe(true)

	session.current = false
	stale.send.mockClear()
	ws.connect()

	expect(stale.close).toHaveBeenCalledTimes(1)
	expect(FakeSocket.instances).toHaveLength(2)
	expect(FakeSocket.instances[1]).not.toBe(stale)
	expect(ws.status.value).toBe('connecting')
	expect(ws.authenticated.value).toBe(false)

	session.current = true
	const current = FakeSocket.instances[1]
	current.onopen?.()
	current.onmessage?.(authSuccessFrame())
	current.send.mockClear()

	ws.subscribe('notification.created', vi.fn())
	expect(current.send).toHaveBeenCalledTimes(1)
	expect(current.send).toHaveBeenCalledWith(JSON.stringify({
		action: 'subscribe',
		event: 'notification.created',
	}))
	expect(stale.send).not.toHaveBeenCalled()
})

it('replaces a socket whose session changed without waiting for a frame, keeping subscriptions', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const stale = FakeSocket.instances[0]
	stale.onopen?.()
	stale.onmessage?.(authSuccessFrame())
	ws.subscribe('timer.created', vi.fn())
	expect(ws.authenticated.value).toBe(true)

	session.current = false
	ws.closeStaleConnection()

	expect(stale.close).toHaveBeenCalledTimes(1)
	expect(FakeSocket.instances).toHaveLength(2)
	expect(ws.status.value).toBe('connecting')
	expect(ws.authenticated.value).toBe(false)

	session.current = true
	const current = FakeSocket.instances[1]
	current.onopen?.()
	current.send.mockClear()
	current.onmessage?.(authSuccessFrame())
	expect(current.send).toHaveBeenCalledTimes(1)
	expect(current.send).toHaveBeenCalledWith(JSON.stringify({
		action: 'subscribe',
		event: 'timer.created',
	}))
})

it('flags possibly missed events only while a dropped connection is pending a reconnect', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	expect(ws.mayHaveMissedEvents.value).toBe(false)

	socket.onclose?.()
	expect(ws.mayHaveMissedEvents.value).toBe(true)

	ws.disconnect()
	expect(ws.mayHaveMissedEvents.value).toBe(false)
})

it('awaits the first connection of a user session until it authenticates', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	expect(ws.isAwaitingFirstConnection()).toBe(true)
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	expect(ws.isAwaitingFirstConnection()).toBe(true)
	socket.onmessage?.(authSuccessFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(false)
})

it('stops awaiting the first connection when it closes before auth', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	FakeSocket.instances[0].onclose?.()
	expect(ws.isAwaitingFirstConnection()).toBe(false)

	vi.runOnlyPendingTimers()
	expect(FakeSocket.instances).toHaveLength(2)
	expect(ws.isAwaitingFirstConnection()).toBe(false)
})

it('stops awaiting the first connection when the socket cannot be created', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', class {
		static OPEN = 1
		static CONNECTING = 0
		constructor() { throw new Error('blocked') }
	})
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	expect(ws.isAwaitingFirstConnection()).toBe(false)
	expect(ws.status.value).toBe('idle')
})

it('does not settle the first connection after disconnecting before the timeout', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	ws.disconnect()
	vi.advanceTimersByTime(5000)
	expect(ws.isAwaitingFirstConnection()).toBe(true)
})

it('stops awaiting the first connection when authentication fails', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	socket.onmessage?.(authErrorFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(false)
})

it('stops awaiting the first connection 5 seconds after connecting if auth never answers', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	FakeSocket.instances[0].onopen?.()
	vi.advanceTimersByTime(4999)
	expect(ws.isAwaitingFirstConnection()).toBe(true)
	vi.advanceTimersByTime(1)
	expect(ws.isAwaitingFirstConnection()).toBe(false)
	expect(ws.status.value).toBe('authenticating')
})

it('awaits the first connection again after the session changes until the new socket authenticates', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const stale = FakeSocket.instances[0]
	stale.onopen?.()
	stale.onmessage?.(authSuccessFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(false)

	session.current = false
	ws.closeStaleConnection()
	session.current = true
	expect(FakeSocket.instances).toHaveLength(2)
	expect(ws.status.value).toBe('connecting')
	expect(ws.isAwaitingFirstConnection()).toBe(true)

	const current = FakeSocket.instances[1]
	current.onopen?.()
	expect(ws.isAwaitingFirstConnection()).toBe(true)
	current.onmessage?.(authSuccessFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(false)
	expect(ws.status.value).toBe('authenticated')
})

it('awaits the first connection again when a frame reveals the session changed', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	const stale = FakeSocket.instances[0]
	stale.onopen?.()
	stale.onmessage?.(authSuccessFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(false)

	session.current = false
	stale.onmessage?.(timerFrame())
	expect(ws.isAwaitingFirstConnection()).toBe(true)
})

it('stops awaiting the re-armed first connection 5 seconds after the session changes if auth never answers', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	FakeSocket.instances[0].onopen?.()
	FakeSocket.instances[0].onmessage?.(authSuccessFrame())

	session.current = false
	ws.closeStaleConnection()
	session.current = true
	FakeSocket.instances[1].onopen?.()
	vi.advanceTimersByTime(4999)
	expect(ws.isAwaitingFirstConnection()).toBe(true)
	vi.advanceTimersByTime(1)
	expect(ws.isAwaitingFirstConnection()).toBe(false)
	expect(ws.status.value).toBe('authenticating')
})

it('leaves a current or absent connection alone', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.closeStaleConnection()
	expect(FakeSocket.instances).toHaveLength(0)

	ws.connect()
	const socket = FakeSocket.instances[0]
	socket.onopen?.()
	ws.closeStaleConnection()
	expect(socket.close).not.toHaveBeenCalled()
	expect(ws.status.value).toBe('authenticating')
})

it('does not open a socket for a link share session', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	auth.tokenType = AUTH_TYPES.LINK_SHARE
	const ws = useWebSocket()
	ws.connect()
	expect(FakeSocket.instances).toHaveLength(0)
	expect(ws.isAwaitingFirstConnection()).toBe(false)
})

it.each([
	['http://localhost', 'ws://localhost/api/v2/ws'],
	['https://x/prefix', 'wss://x/prefix/api/v2/ws'],
	['https://x/prefix/api/v1', 'wss://x/prefix/api/v2/ws'],
	['', 'ws://localhost:3000/api/v2/ws'],
])('connects to the v2 socket under %o', (apiUrl, socketUrl) => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = apiUrl
	useWebSocket().connect()
	expect(FakeSocket.instances.map(socket => socket.url)).toEqual([socketUrl])
})

it('backs off further on each connection that closes before authenticating, and resets after auth succeeds', () => {
	vi.useFakeTimers()
	vi.spyOn(Math, 'random').mockReturnValue(0.5)
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	useWebSocket().connect()

	for (const delay of [1000, 2000, 4000]) {
		const socket = FakeSocket.instances[FakeSocket.instances.length - 1]
		const count = FakeSocket.instances.length
		socket.onopen?.()
		socket.onclose?.()
		vi.advanceTimersByTime(delay - 1)
		expect(FakeSocket.instances).toHaveLength(count)
		vi.advanceTimersByTime(1)
		expect(FakeSocket.instances).toHaveLength(count + 1)
	}

	const socket = FakeSocket.instances[FakeSocket.instances.length - 1]
	const count = FakeSocket.instances.length
	socket.onopen?.()
	socket.onmessage?.(authSuccessFrame())
	socket.onclose?.()
	vi.advanceTimersByTime(1000)
	expect(FakeSocket.instances).toHaveLength(count + 1)
})

it('restarts the backoff after a terminal auth failure', () => {
	vi.useFakeTimers()
	vi.spyOn(Math, 'random').mockReturnValue(0.5)
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	const ws = useWebSocket()
	ws.connect()
	FakeSocket.instances[0].onclose?.()
	vi.advanceTimersByTime(1000)
	const rejected = FakeSocket.instances[1]
	rejected.onopen?.()
	rejected.onmessage?.(authErrorFrame())

	ws.connect()
	FakeSocket.instances[2].onclose?.()
	vi.advanceTimersByTime(1000)
	expect(FakeSocket.instances).toHaveLength(4)
})

it('re-authenticates the open socket with the refreshed token', () => {
	const {ws, socket} = authenticatedSocket()

	tokenRefreshed('fresh')

	expect(socket.send).toHaveBeenCalledTimes(1)
	expect(socket.send).toHaveBeenCalledWith(JSON.stringify({action: 'auth', token: 'fresh'}))
	expect(FakeSocket.instances).toHaveLength(1)
	expect(ws.status.value).toBe('authenticated')
})

it('does not resubscribe when a re-auth succeeds', () => {
	const {ws, socket} = authenticatedSocket()
	ws.subscribe('timer.created', vi.fn())
	tokenRefreshed('fresh')
	socket.send.mockClear()

	socket.onmessage?.(authSuccessFrame())

	expect(socket.send).not.toHaveBeenCalled()
	expect(ws.status.value).toBe('authenticated')
})

it('does not re-authenticate while not yet authenticated', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost'
	tokenRefreshed('fresh')
	expect(FakeSocket.instances).toHaveLength(0)

	useWebSocket().connect()
	const socket = FakeSocket.instances[0]
	tokenRefreshed('fresher')
	socket.onopen?.()
	tokenRefreshed('freshest')

	expect(socket.send).toHaveBeenCalledTimes(1)
	expect(socket.send).toHaveBeenCalledWith(JSON.stringify({action: 'auth', token: 'fresher'}))
})

it('does not re-authenticate with a non-user token', () => {
	const {socket} = authenticatedSocket()
	auth.tokenType = AUTH_TYPES.LINK_SHARE

	tokenRefreshed('share')

	expect(socket.send).not.toHaveBeenCalled()
})

it('stops the socket when the server rejects a re-auth', () => {
	vi.useFakeTimers()
	const {ws, socket} = authenticatedSocket()
	tokenRefreshed('fresh')

	socket.onmessage?.(authErrorFrame())

	vi.runOnlyPendingTimers()
	expect(socket.close).toHaveBeenCalled()
	expect(FakeSocket.instances).toHaveLength(1)
	expect(ws.status.value).toBe('idle')
})
