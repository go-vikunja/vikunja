import {describe, expect, it, vi} from 'vitest'
import {effectScope, nextTick} from 'vue'
import {ApiError} from '@/client/problem'
import {useAccessLossConfirmation} from './useAccessLossConfirmation'

function conflict() {
	return new ApiError({status: 409, detail: 'conflict', code: 3018, i18n_params: {webhooks: '2', link_shares: '1'}})
}

function setup() {
	const scope = effectScope()
	const confirmation = scope.run(() => useAccessLossConfirmation())!
	return {
		...confirmation,
		scope,
	}
}

function pendingAttempt() {
	let reject!: (error: unknown) => void
	const promise = new Promise<unknown>((_, rejectPromise) => {
		reject = rejectPromise
	})
	return {promise, reject}
}

describe('useAccessLossConfirmation', () => {
	it('asks the server first and skips the dialog without a conflict', async () => {
		const {run, counts} = setup()
		const attempt = vi.fn().mockResolvedValue({})
		expect(await run(attempt)).toBe(true)
		expect(attempt).toHaveBeenCalledExactlyOnceWith({ask_before_access_loss: true})
		expect(counts.value).toBeNull()
	})

	it('retries with the chosen cleanup after a conflict', async () => {
		const {run, counts, cleanup, confirm} = setup()
		const attempt = vi.fn().mockRejectedValueOnce(conflict()).mockResolvedValueOnce({})
		const result = run(attempt)
		await nextTick()
		await nextTick()
		expect(counts.value).toEqual({webhooks: 2, linkShares: 1})
		cleanup.value = {keep_webhooks: true, revoke_link_shares: true}
		confirm()
		expect(await result).toBe(true)
		expect(attempt).toHaveBeenLastCalledWith({keep_webhooks: true, revoke_link_shares: true})
		expect(counts.value).toBeNull()
	})

	it('does not retry when cancelled', async () => {
		const {run, cancel} = setup()
		const attempt = vi.fn().mockRejectedValueOnce(conflict())
		const result = run(attempt)
		await nextTick()
		await nextTick()
		cancel()
		expect(await result).toBe(false)
		expect(attempt).toHaveBeenCalledOnce()
	})

	it('cancels the open dialog when run again', async () => {
		const {run, confirm} = setup()
		const first = vi.fn().mockRejectedValueOnce(conflict())
		const firstResult = run(first)
		await nextTick()
		await nextTick()
		const second = vi.fn().mockRejectedValueOnce(conflict()).mockResolvedValueOnce({})
		const secondResult = run(second)
		expect(await firstResult).toBe(false)
		await nextTick()
		await nextTick()
		confirm()
		expect(await secondResult).toBe(true)
		expect(first).toHaveBeenCalledOnce()
		expect(second).toHaveBeenCalledTimes(2)
	})

	it('cancels the open dialog when its scope is disposed', async () => {
		const {run, scope} = setup()
		const attempt = vi.fn().mockRejectedValueOnce(conflict())
		const result = run(attempt)
		await nextTick()
		await nextTick()
		scope.stop()
		expect(await result).toBe(false)
		expect(attempt).toHaveBeenCalledOnce()
	})

	it('ignores an older conflict arriving after the newer dialog opens', async () => {
		const {run, counts, cleanup, confirm} = setup()
		const firstRequest = pendingAttempt()
		const secondRequest = pendingAttempt()
		const first = vi.fn().mockReturnValueOnce(firstRequest.promise)
		const second = vi.fn().mockReturnValueOnce(secondRequest.promise).mockResolvedValueOnce({})
		const firstResult = run(first)
		const secondResult = run(second)
		secondRequest.reject(conflict())
		await nextTick()
		cleanup.value = {keep_webhooks: true, revoke_link_shares: true}

		firstRequest.reject(new ApiError({status: 409, detail: 'conflict', code: 3018, i18n_params: {webhooks: '9', link_shares: '8'}}))
		await nextTick()
		expect(counts.value).toEqual({webhooks: 2, linkShares: 1})
		expect(cleanup.value).toEqual({keep_webhooks: true, revoke_link_shares: true})
		expect(await firstResult).toBe(false)
		confirm()
		expect(await secondResult).toBe(true)
		expect(first).toHaveBeenCalledOnce()
		expect(second).toHaveBeenLastCalledWith({keep_webhooks: true, revoke_link_shares: true})
		expect(counts.value).toBeNull()
	})

	it('settles a late conflict after its scope is disposed', async () => {
		const {run, counts, scope} = setup()
		const request = pendingAttempt()
		const attempt = vi.fn().mockReturnValueOnce(request.promise)
		const result = run(attempt)
		scope.stop()
		request.reject(conflict())
		await nextTick()
		expect(counts.value).toBeNull()
		expect(await result).toBe(false)
		expect(attempt).toHaveBeenCalledOnce()
	})

	it('does not retry a confirmation when its scope is disposed before resuming', async () => {
		const {run, confirm, scope} = setup()
		const attempt = vi.fn().mockRejectedValueOnce(conflict()).mockResolvedValueOnce({})
		const result = run(attempt)
		await nextTick()
		confirm()
		scope.stop()
		expect(await result).toBe(false)
		expect(attempt).toHaveBeenCalledOnce()
	})

	it('rethrows other errors', async () => {
		const {run} = setup()
		const error = new ApiError({status: 403, detail: 'forbidden', code: 3001})
		await expect(run(vi.fn().mockRejectedValue(error))).rejects.toBe(error)
	})
})
