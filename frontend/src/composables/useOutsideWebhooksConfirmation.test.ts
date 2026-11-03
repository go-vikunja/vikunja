import {effectScope} from 'vue'
import {describe, expect, it, vi} from 'vitest'

import {ApiError} from '@/client/problem'
import {useOutsideWebhooksConfirmation} from './useOutsideWebhooksConfirmation'

const conflict = () => new ApiError({status: 409, detail: 'confirm', code: 3018, i18n_params: {count: '2'}})

function setup() {
	const scope = effectScope()
	const confirmation = scope.run(() => useOutsideWebhooksConfirmation())!
	return {scope, ...confirmation}
}

describe('useOutsideWebhooksConfirmation', () => {
	it('sends no decision first and never prompts when the request succeeds', async () => {
		const {prompt, run} = setup()
		const request = vi.fn(async () => {})

		await expect(run(request)).resolves.toBe(true)
		expect(request).toHaveBeenCalledOnce()
		expect(request).toHaveBeenCalledWith()
		expect(prompt.value).toBeNull()
	})

	it.each([true, false])('retries with the chosen decision %s', async choice => {
		const {prompt, run} = setup()
		const request = vi.fn(async (clear?: boolean) => {
			if (clear === undefined) throw conflict()
		})

		const result = run(request)
		await vi.waitFor(() => expect(prompt.value?.count).toBe(2))
		prompt.value!.answer(choice)

		await expect(result).resolves.toBe(true)
		expect(request).toHaveBeenLastCalledWith(choice)
		expect(prompt.value).toBeNull()
	})

	it('does not retry when the prompt is cancelled', async () => {
		const {prompt, run} = setup()
		const request = vi.fn(async () => { throw conflict() })

		const result = run(request)
		await vi.waitFor(() => expect(prompt.value).not.toBeNull())
		prompt.value!.answer(null)

		await expect(result).resolves.toBe(false)
		expect(request).toHaveBeenCalledOnce()
	})

	it('rethrows other errors without prompting', async () => {
		const {prompt, run} = setup()
		const cause = new ApiError({status: 409, detail: 'already shared', code: 3007})

		await expect(run(async () => { throw cause })).rejects.toBe(cause)
		expect(prompt.value).toBeNull()
	})

	it('cancels a pending prompt when a new one opens or the scope ends', async () => {
		const {scope, prompt, run} = setup()
		const first = run(async () => { throw conflict() })
		await vi.waitFor(() => expect(prompt.value).not.toBeNull())
		const firstPrompt = prompt.value

		const second = run(async () => { throw conflict() })
		await expect(first).resolves.toBe(false)
		await vi.waitFor(() => expect(prompt.value).not.toBe(firstPrompt))

		scope.stop()
		await expect(second).resolves.toBe(false)
	})
})
