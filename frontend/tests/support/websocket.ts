import WebSocket from 'ws'

const API_URL = process.env.API_URL || 'http://localhost:3456/api/v1'

export interface WsMessage {
	event?: string
	action?: string
	success?: boolean
	error?: string
	topic?: string
	data?: unknown
}

/**
 * Returns the WebSocket URL derived from the API base URL.
 */
export function getWsUrl(): string {
	return API_URL.replace(/\/+$/, '').replace(/^http/, 'ws') + '/ws'
}

/**
 * Opens a raw WebSocket connection to the API.
 */
export function openWs(): Promise<WebSocket> {
	return new Promise((resolve, reject) => {
		const ws = new WebSocket(getWsUrl())
		ws.on('open', () => resolve(ws))
		ws.on('error', reject)
	})
}

/**
 * Waits for the next message on a WebSocket connection.
 */
export function waitForMessage(ws: WebSocket, timeout = 5000): Promise<WsMessage> {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(() => reject(new Error('WebSocket message timeout')), timeout)
		ws.once('message', (data) => {
			clearTimeout(timer)
			resolve(JSON.parse(data.toString()))
		})
	})
}

/**
 * Sends a JSON message on the WebSocket.
 */
export function sendMessage(ws: WebSocket, msg: object): void {
	ws.send(JSON.stringify(msg))
}

/**
 * Authenticates a WebSocket connection and returns the auth.success message.
 */
export async function authenticateWs(ws: WebSocket, token: string): Promise<WsMessage> {
	sendMessage(ws, {action: 'auth', token})
	const msg = await waitForMessage(ws)
	if (msg.action !== 'auth.success') {
		throw new Error(`Expected auth.success, got: ${JSON.stringify(msg)}`)
	}
	return msg
}

/**
 * Subscribes to a topic on an authenticated WebSocket connection.
 */
export function subscribeWs(ws: WebSocket, topic: string): void {
	sendMessage(ws, {action: 'subscribe', topic})
}

/**
 * Collects all messages received within a time window.
 */
export function collectMessages(ws: WebSocket, duration: number): Promise<WsMessage[]> {
	return new Promise((resolve) => {
		const messages: WsMessage[] = []
		const handler = (data: WebSocket.Data) => {
			messages.push(JSON.parse(data.toString()))
		}
		ws.on('message', handler)
		setTimeout(() => {
			ws.off('message', handler)
			resolve(messages)
		}, duration)
	})
}

/**
 * Closes a WebSocket connection safely.
 */
export function closeWs(ws: WebSocket): void {
	if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
		ws.close()
	}
}
