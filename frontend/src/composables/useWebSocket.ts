import {ref, readonly} from 'vue'

import {getToken} from '@/helpers/auth'

type MessageCallback = (msg: WebSocketEvent) => void

interface WebSocketEvent {
	event?: string
	action?: string
	success?: boolean
	error?: string
	data?: unknown
}

const RECONNECT_BASE_DELAY = 1000
const RECONNECT_MAX_DELAY = 30000

let socket: WebSocket | null = null
let reconnectAttempt = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
const subscriptions = new Map<string, Set<MessageCallback>>()
const connected = ref(false)
const authenticated = ref(false)
let manuallyDisconnected = false

function getWebSocketUrl(): string {
	const base = window.API_URL.replace(/\/+$/, '')
	const wsProtocol = base.startsWith('https') ? 'wss' : 'ws'
	return base.replace(/^https?/, wsProtocol) + '/ws'
}

function sendMessage(msg: object) {
	if (socket?.readyState === WebSocket.OPEN) {
		socket.send(JSON.stringify(msg))
	}
}

function sendAuth() {
	const token = getToken()
	if (token) {
		sendMessage({action: 'auth', token})
	}
}

function resubscribeAll() {
	for (const event of subscriptions.keys()) {
		sendMessage({action: 'subscribe', event})
	}
}

function handleMessage(event: MessageEvent) {
	let msg: WebSocketEvent
	try {
		msg = JSON.parse(event.data)
	} catch {
		console.warn('WebSocket: invalid message', event.data)
		return
	}

	// Handle auth success
	if (msg.action === 'auth.success' && msg.success) {
		authenticated.value = true
		console.debug('WebSocket: authenticated')
		resubscribeAll()
		return
	}

	// Handle auth error - treat as terminal (no reconnect) so we don't
	// thrash the WS endpoint with a bad token. Fallback polling kicks in.
	if (msg.error === 'invalid_token' || msg.error === 'auth_required') {
		console.warn('WebSocket: auth failed:', msg.error)
		manuallyDisconnected = true
		authenticated.value = false
		connected.value = false
		socket?.close()
		socket = null
		return
	}

	// Handle regular events — route by event name
	if (msg.event) {
		const callbacks = subscriptions.get(msg.event)
		if (callbacks) {
			for (const cb of callbacks) {
				cb(msg)
			}
		}
	}
}

function scheduleReconnect() {
	if (manuallyDisconnected) {
		return
	}

	if (reconnectTimer) {
		clearTimeout(reconnectTimer)
		reconnectTimer = null
	}

	const delay = Math.min(
		RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttempt),
		RECONNECT_MAX_DELAY,
	)
	reconnectAttempt++
	console.debug(`WebSocket: reconnecting in ${delay}ms (attempt ${reconnectAttempt})`)

	reconnectTimer = setTimeout(() => {
		reconnectTimer = null
		connect()
	}, delay)
}

function connect() {
	if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
		return
	}

	const token = getToken()
	if (!token) {
		return
	}

	manuallyDisconnected = false
	authenticated.value = false
	const url = getWebSocketUrl()

	try {
		socket = new WebSocket(url)
	} catch (e) {
		console.warn('WebSocket: failed to create connection', e)
		scheduleReconnect()
		return
	}

	socket.onopen = () => {
		connected.value = true
		reconnectAttempt = 0
		console.debug('WebSocket: connected, sending auth')
		sendAuth()
	}

	socket.onmessage = handleMessage

	socket.onclose = () => {
		connected.value = false
		authenticated.value = false
		socket = null
		scheduleReconnect()
	}

	socket.onerror = () => {
		// onclose will fire after onerror, which handles reconnect
	}
}

function disconnect() {
	manuallyDisconnected = true
	if (reconnectTimer) {
		clearTimeout(reconnectTimer)
		reconnectTimer = null
	}
	reconnectAttempt = 0
	if (socket) {
		socket.close()
		socket = null
	}
	connected.value = false
	authenticated.value = false
	subscriptions.clear()
}

function subscribe(event: string, callback: MessageCallback): () => void {
	if (!subscriptions.has(event)) {
		subscriptions.set(event, new Set())
	}
	subscriptions.get(event)!.add(callback)

	// Only send subscribe if already authenticated
	// (otherwise it will be sent after auth succeeds)
	if (authenticated.value) {
		sendMessage({action: 'subscribe', event})
	}

	return () => {
		const callbacks = subscriptions.get(event)
		if (callbacks) {
			callbacks.delete(callback)
			if (callbacks.size === 0) {
				subscriptions.delete(event)
				sendMessage({action: 'unsubscribe', event})
			}
		}
	}
}

export function useWebSocket() {
	return {
		connect,
		disconnect,
		subscribe,
		connected: readonly(connected),
		authenticated: readonly(authenticated),
	}
}
