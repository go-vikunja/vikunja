import {ApiError} from '@/client/problem'

const OUTSIDE_WEBHOOKS_CONFLICT_CODE = 3018

// How many webhooks the server wants a decision on before deleting them, or null for any other error.
export function outsideWebhooksToConfirm(cause: unknown): number | null {
	if (!(cause instanceof ApiError) || cause.status !== 409 || cause.code !== OUTSIDE_WEBHOOKS_CONFLICT_CODE) {
		return null
	}
	return Number(cause.i18n_params?.count ?? 0)
}

export const toastUnlessOutsideWebhooksConflict = (_input: unknown, cause: unknown) => outsideWebhooksToConfirm(cause) === null
