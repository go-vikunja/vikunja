import {computed, ref, readonly} from 'vue'

import {getToken, getTokenType, isTokenExpired, refreshToken, RefreshTokenError} from '@/helpers/auth'
import {getApiBaseUrl} from '@/helpers/apiUrl'
import {AUTH_TYPES} from '@/constants/auth'
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
const FIRST_CONNECTION_TIMEOUT = 5000

let socket: WebSocket | null = null
let socketContext: ClientRequestContext | null = null
let reconnectAttempt = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
const subscriptions = new Map<string, Set<MessageCallback>>()
const status = ref<'idle' | 'connecting' | 'authenticating' | 'authenticated'>('idle')
const authenticated = computed(() => status.value === 'authenticated')
const mayHaveMissedEvents = ref(false)
const subscribedAt = ref(0)
let manuallyDisconnected = false
const firstConnectionSettled = ref(false)
let firstConnectionTimer: ReturnType<typeof setTimeout> | undefined

function setFirstConnectionSettled(settled: boolean) {
	firstConnectionSettled.value = settled
	clearTimeout(firstConnectionTimer)
	firstConnectionTimer = undefined
}

function getWebSocketUrl(): string {
	const url = new URL(`${getApiBaseUrl()}/ws`, window.location.origin)
	url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
	return url.href
}

function sendMessage(msg: object) {
	if (socket?.readyState === WebSocket.OPEN) {
		socket.send(JSON.stringify(msg))
	}
}

async function sendAuth(connection: WebSocket) {
	// The server closes the socket on token expiry, so reconnects start stale.
	if (isTokenExpired(getToken())) {
		try {
			await refreshToken(true)
		} catch (e) {
			// A rejected refresh means the session is gone; the stale token's invalid_token stops retries.
			if (!(e instanceof RefreshTokenError && e.failure.kind === 'rejected')) {
				connection.close()
				return
			}
		}
	}
	const token = getToken()
	if (token && socket === connection) {
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
	status.value = 'idle'
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
		status.value = 'authenticated'
		reconnectAttempt = 0
		console.debug('WebSocket: authenticated')
		resubscribeAll()
		// The server never acks a subscribe, so the send time is the earliest point events can reach us.
		subscribedAt.value = Date.now()
		setFirstConnectionSettled(true)
		return
	}

	// Handle auth error - treat as terminal (no reconnect) so we don't
	// thrash the WS endpoint with a bad token. Fallback polling kicks in.
	if (msg.error === 'invalid_token' || msg.error === 'auth_required') {
		console.warn('WebSocket: auth failed:', msg.error)
		manuallyDisconnected = true
		reconnectAttempt = 0
		closeSocket()
		setFirstConnectionSettled(true)
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

// Not a computed: the token isn't reactive, so it would cache a link share's false across login.
function isAwaitingFirstConnection(): boolean {
	return !firstConnectionSettled.value && mayOpenSocket()
}

function connect() {
	// A connection stays authenticated as whoever opened it, so a session change must tear it down, never adopt it.
	if (socketContext && !isClientRequestContextCurrent(socketContext)) {
		closeSocket()
		setFirstConnectionSettled(false)
	}

	if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
		return
	}

	if (!mayOpenSocket()) {
		return
	}

	manuallyDisconnected = false
	const url = getWebSocketUrl()

	const context = captureClientRequestContext()
	try {
		socket = new WebSocket(url)
	} catch (e) {
		console.warn('WebSocket: failed to create connection', e)
		status.value = 'idle'
		setFirstConnectionSettled(true)
		scheduleReconnect()
		return
	}
	socketContext = context
	status.value = 'connecting'
	if (!firstConnectionSettled.value && !firstConnectionTimer) {
		firstConnectionTimer = setTimeout(() => setFirstConnectionSettled(true), FIRST_CONNECTION_TIMEOUT)
	}

	const connection = socket
	const isCurrent = () => socket === connection && isClientRequestContextCurrent(context)

	function dropStaleConnection() {
		if (socket !== connection) {
			connection.close()
			return
		}
		closeSocket()
		setFirstConnectionSettled(false)
		scheduleReconnect()
	}

	socket.onopen = () => {
		if (!isCurrent()) {
			dropStaleConnection()
			return
		}
		status.value = 'authenticating'
		console.debug('WebSocket: connected, sending auth')
		void sendAuth(connection)
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
		setFirstConnectionSettled(true)
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
	setFirstConnectionSettled(false)
	// Nothing else reconnects: ContentAuth connects only on mount.
	connect()
}

function disconnect() {
	manuallyDisconnected = true
	reconnectAttempt = 0
	closeSocket()
	setFirstConnectionSettled(false)
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
		status: readonly(status),
		authenticated,
		isAwaitingFirstConnection,
		mayHaveMissedEvents: readonly(mayHaveMissedEvents),
		subscribedAt: readonly(subscribedAt),
	}
}
