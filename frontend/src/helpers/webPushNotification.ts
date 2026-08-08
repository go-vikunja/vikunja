export interface WebPushPayload {
	title: string
	body: string
	url: string
	tag: string
	notification_id?: number
	test?: boolean
}

export function webPushNotificationTarget(
	payload: Pick<WebPushPayload, 'url' | 'notification_id'>,
	scope: string,
	origin: string,
): string {
	const target = new URL(payload.url.replace(/^\//, ''), scope)
	if (target.origin !== origin) {
		return scope
	}
	if (payload.notification_id) {
		target.searchParams.set('vikunja_notification', String(payload.notification_id))
	}
	return target.toString()
}
