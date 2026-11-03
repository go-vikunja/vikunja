import {computed, onScopeDispose, shallowRef} from 'vue'
import {outsideWebhooksToConfirm} from '@/client/queries/outsideWebhooks'

export type OutsideWebhooksPrompt = {
	count: number
	answer: (clearOutsideWebhooks: boolean | null) => void
}

export function useOutsideWebhooksConfirmation() {
	const current = shallowRef<OutsideWebhooksPrompt | null>(null)

	function ask(count: number) {
		current.value?.answer(null)
		return new Promise<boolean | null>(resolve => {
			const prompt: OutsideWebhooksPrompt = {
				count,
				answer: choice => {
					if (current.value === prompt) current.value = null
					resolve(choice)
				},
			}
			current.value = prompt
		})
	}

	// Resolves false when the user cancelled at the prompt.
	async function run(request: (clearOutsideWebhooks?: boolean) => Promise<unknown>): Promise<boolean> {
		let count: number | null
		try {
			await request()
			return true
		} catch (e) {
			count = outsideWebhooksToConfirm(e)
			if (count === null) throw e
		}
		const choice = await ask(count)
		if (choice === null) return false
		await request(choice)
		return true
	}

	onScopeDispose(() => current.value?.answer(null))

	return {
		prompt: computed(() => current.value),
		run,
	}
}
