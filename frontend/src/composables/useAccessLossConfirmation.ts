import {onScopeDispose, ref} from 'vue'
import {ApiError} from '@/client/problem'
import {defaultAccessLossCleanup, type AccessLossChoice} from '@/client/queries/accessLossCleanup'

const ERR_CODE_ACCESS_LOSS_NEEDS_CONFIRMATION = 3022

export type AccessLossCounts = {
	webhooks: number
	linkShares: number
}

function countsFrom(error: unknown): AccessLossCounts | null {
	if (!(error instanceof ApiError) || error.code !== ERR_CODE_ACCESS_LOSS_NEEDS_CONFIRMATION) return null
	return {
		webhooks: Number(error.i18n_params?.webhooks ?? 0),
		linkShares: Number(error.i18n_params?.link_shares ?? 0),
	}
}

export function useAccessLossConfirmation() {
	const counts = ref<AccessLossCounts | null>(null)
	const cleanup = ref(defaultAccessLossCleanup())
	let decide: ((confirmed: boolean) => void) | null = null
	let invocation = 0

	async function run(attempt: (choice: AccessLossChoice) => Promise<unknown>): Promise<boolean> {
		const currentInvocation = ++invocation
		settle(false)
		try {
			await attempt({ask_before_access_loss: true})
			return true
		} catch (e) {
			const conflict = countsFrom(e)
			if (!conflict) throw e
			if (currentInvocation !== invocation) return false
			counts.value = conflict
		}

		cleanup.value = defaultAccessLossCleanup()
		const confirmed = await new Promise<boolean>(resolve => {
			decide = resolve
		})
		if (!confirmed || currentInvocation !== invocation) return false
		await attempt(cleanup.value)
		return true
	}

	function settle(confirmed: boolean) {
		counts.value = null
		decide?.(confirmed)
		decide = null
	}

	onScopeDispose(() => {
		invocation++
		settle(false)
	})

	return {
		counts,
		cleanup,
		run,
		confirm: () => settle(true),
		cancel: () => settle(false),
	}
}
