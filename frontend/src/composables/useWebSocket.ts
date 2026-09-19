import {ref, readonly} from 'vue'

import {getToken, getTokenType} from '@/helpers/auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'
import {
	captureClientRequestContext,
	isClientRequestContextCurrent,
	type ClientRequestContext,
} from '@/client/requestContext'

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
let socketContext: ClientRequestContext | null = null
let reconnectAttempt = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
const subscriptions = new Map<string, Set<MessageCallback>>()
const connected = ref(false)
const authenticated = ref(false)
const mayHaveMissedEvents = ref(false)
const subscribedAt = ref(0)
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

function closeSocket() {
	socket?.close()
	socket = null
	socketContext = null
	connected.value = false
	authenticated.value = false
	mayHaveMissedEvents.value = false
	subscribedAt.value = 0
	if (reconnectTimer) {
		clearTimeout(reconnectTimer)
		reconnectTimer = null
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
		// The server never acks a subscribe, so the send time is the earliest point events can reach us.
		subscribedAt.value = Date.now()
		return
	}

	// Handle auth error - treat as terminal (no reconnect) so we don't
	// thrash the WS endpoint with a bad token. Fallback polling kicks in.
	if (msg.error === 'invalid_token' || msg.error === 'auth_required') {
		console.warn('WebSocket: auth failed:', msg.error)
		manuallyDisconnected = true
		closeSocket()
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
	mayHaveMissedEvents.value = true

	if (reconnectTimer) {
		clearTimeout(reconnectTimer)
		reconnectTimer = null
	}

	const baseDelay = Math.min(
		RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttempt),
		RECONNECT_MAX_DELAY,
	)
	// Add ±25% jitter to prevent thundering herd on server restart
	const jitter = baseDelay * (0.75 + Math.random() * 0.5)
	const delay = Math.round(jitter)
	reconnectAttempt++
	console.debug(`WebSocket: reconnecting in ${delay}ms (attempt ${reconnectAttempt})`)

	reconnectTimer = setTimeout(() => {
		reconnectTimer = null
		connect()
	}, delay)
}

// Link share tokens are rejected by the socket, so their session never opens one.
function mayOpenSocket(): boolean {
	return getTokenType(getToken()) === AUTH_TYPES.USER
}

function connect() {
	// A connection stays authenticated as whoever opened it, so a session change must tear it down, never adopt it.
	if (socketContext && !isClientRequestContextCurrent(socketContext)) {
		closeSocket()
	}

	if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
		return
	}

	if (!mayOpenSocket()) {
		return
	}

	manuallyDisconnected = false
	authenticated.value = false
	const url = getWebSocketUrl()

	const context = captureClientRequestContext()
	try {
		socket = new WebSocket(url)
	} catch (e) {
		console.warn('WebSocket: failed to create connection', e)
		scheduleReconnect()
		return
	}
	socketContext = context

	const connection = socket
	const isCurrent = () => socket === connection && isClientRequestContextCurrent(context)

	function dropStaleConnection() {
		if (socket !== connection) {
			connection.close()
			return
		}
		closeSocket()
		scheduleReconnect()
	}

	socket.onopen = () => {
		if (!isCurrent()) {
			dropStaleConnection()
			return
		}
		connected.value = true
		reconnectAttempt = 0
		console.debug('WebSocket: connected, sending auth')
		sendAuth()
	}

	socket.onmessage = event => {
		if (!isCurrent()) {
			dropStaleConnection()
			return
		}
		handleMessage(event)
	}

	socket.onclose = () => {
		if (!isCurrent()) {
			dropStaleConnection()
			return
		}
		closeSocket()
		scheduleReconnect()
	}

	socket.onerror = () => {
		// onclose will fire after onerror, which handles reconnect
	}
}

// Eager counterpart to the lazy fence in connect(): an in-tab identity change must not leave the
// previous session's authenticated socket receiving frames until one happens to arrive.
function closeStaleConnection() {
	if (!socketContext || isClientRequestContextCurrent(socketContext)) {
		return
	}
	closeSocket()
}

function disconnect() {
	manuallyDisconnected = true
	reconnectAttempt = 0
	closeSocket()
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
		closeStaleConnection,
		subscribe,
		connected: readonly(connected),
		authenticated: readonly(authenticated),
		mayHaveMissedEvents: readonly(mayHaveMissedEvents),
		subscribedAt: readonly(subscribedAt),
	}
}
