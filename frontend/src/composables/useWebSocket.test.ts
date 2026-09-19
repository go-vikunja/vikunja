import {
	afterEach,
	expect,
	it,
	vi,
} from 'vitest'
import {useWebSocket} from './useWebSocket'
import {AUTH_TYPES} from '@/modelTypes/IUser'
const auth = vi.hoisted(() => ({tokenType: 1}))
vi.mock('@/helpers/auth', () => ({
	getToken: () => 'token',
	getTokenType: () => auth.tokenType,
}))
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
	constructor() { FakeSocket.instances.push(this) }
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

afterEach(() => {
	useWebSocket().disconnect()
	vi.useRealTimers()
	vi.unstubAllGlobals()
	FakeSocket.instances = []
	session.current = true
	auth.tokenType = AUTH_TYPES.USER
})

it('routes messages from the current connection to subscribers', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost/api/v1'
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
	window.API_URL = 'http://localhost/api/v1'
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
	expect(ws.connected.value).toBe(true)
})

it('closes the socket and reconnects after the authenticated session changes', () => {
	vi.useFakeTimers()
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost/api/v1'
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
	expect(ws.connected.value).toBe(false)
	expect(ws.authenticated.value).toBe(false)

	session.current = true
	vi.runOnlyPendingTimers()
	expect(FakeSocket.instances).toHaveLength(2)
	expect(FakeSocket.instances[1]).not.toBe(stale)
})

it('tears down a socket opened by a previous session instead of reusing it', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost/api/v1'
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
	expect(ws.connected.value).toBe(false)
	expect(ws.authenticated.value).toBe(false)

	ws.subscribe('notification.created', vi.fn())
	expect(stale.send).not.toHaveBeenCalled()
})

it('does not open a socket for a link share session', () => {
	vi.stubGlobal('WebSocket', FakeSocket)
	window.API_URL = 'http://localhost/api/v1'
	auth.tokenType = AUTH_TYPES.LINK_SHARE
	useWebSocket().connect()
	expect(FakeSocket.instances).toHaveLength(0)
})
