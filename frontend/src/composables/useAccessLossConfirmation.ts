import {ref} from 'vue'
import {ApiError} from '@/client/problem'
import {defaultAccessLossCleanup, type AccessLossChoice} from '@/client/queries/accessLossCleanup'

const ERR_CODE_ACCESS_LOSS_NEEDS_CONFIRMATION = 3018

export type AccessLossCounts = {webhooks: number, linkShares: number}

function countsFrom(error: unknown): AccessLossCounts | null {
	if (!(error instanceof ApiError) || error.code !== ERR_CODE_ACCESS_LOSS_NEEDS_CONFIRMATION) return null
	return {
		webhooks: Number(error.i18n_params?.webhooks ?? 0),
		linkShares: Number(error.i18n_params?.link_shares ?? 0),
	}
}

// Tries without asking first; only a 409 from the server opens the dialog.
export function useAccessLossConfirmation() {
	const counts = ref<AccessLossCounts | null>(null)
	const cleanup = ref(defaultAccessLossCleanup())
	let decide: ((confirmed: boolean) => void) | null = null

	async function run(attempt: (choice: AccessLossChoice) => Promise<unknown>): Promise<boolean> {
		try {
			await attempt({confirm_access_loss: true})
			return true
		} catch (e) {
			const conflict = countsFrom(e)
			if (!conflict) throw e
			counts.value = conflict
		}

		cleanup.value = defaultAccessLossCleanup()
		const confirmed = await new Promise<boolean>(resolve => {
			decide = resolve
		})
		counts.value = null
		if (!confirmed) return false
		await attempt(cleanup.value)
		return true
	}

	function settle(confirmed: boolean) {
		decide?.(confirmed)
		decide = null
	}

	return {
		counts,
		cleanup,
		run,
		confirm: () => settle(true),
		cancel: () => settle(false),
	}
}
