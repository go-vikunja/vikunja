import {afterEach, expect, it, vi} from 'vitest'
import {useWebSocket} from './useWebSocket'
vi.mock('@/helpers/auth', () => ({getToken: () => 'token'}))
vi.mock('@/client/requestContext', () => ({captureClientRequestContext: () => ({}), isClientRequestContextCurrent: () => true}))
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
	close() { this.readyState = 3 }
	constructor() { FakeSocket.instances.push(this) }
}

afterEach(() => {
	useWebSocket().disconnect()
	vi.unstubAllGlobals()
	FakeSocket.instances = []
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
	old.onmessage?.(new MessageEvent('message', {data: JSON.stringify({event: 'timer.created', data: {id: 1}})}))
	old.onclose?.()
	expect(onTimer).not.toHaveBeenCalled()
	expect(ws.connected.value).toBe(true)
})
