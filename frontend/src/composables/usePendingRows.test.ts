import {describe, expect, it} from 'vitest'
import {usePendingRows} from './usePendingRows'

function deferred() {
	let resolve!: () => void
	let reject!: (cause: Error) => void
	const promise = new Promise<void>((res, rej) => {
		resolve = res
		reject = rej
	})
	return {
		promise,
		resolve,
		reject,
	}
}

describe('usePendingRows', () => {
	it('marks a row pending until its mutation resolves', async () => {
		const rows = usePendingRows()
		const request = deferred()
		const run = rows.run({mutateAsync: () => request.promise}, 3)
		expect(rows.has(3)).toBe(true)
		expect(rows.has(4)).toBe(false)
		request.resolve()
		await expect(run).resolves.toBe(true)
		expect(rows.has(3)).toBe(false)
	})

	it('clears the row and swallows the rejection when the mutation fails', async () => {
		const rows = usePendingRows()
		const request = deferred()
		const run = rows.run({mutateAsync: () => request.promise}, 3)
		expect(rows.has(3)).toBe(true)
		request.reject(new Error('failed'))
		await expect(run).resolves.toBe(false)
		expect(rows.has(3)).toBe(false)
	})

	it('ignores a second run while the row is pending', async () => {
		const rows = usePendingRows()
		const request = deferred()
		const first = rows.run({mutateAsync: () => request.promise}, 3)
		await expect(rows.run({mutateAsync: () => request.promise}, 3)).resolves.toBe(false)
		request.resolve()
		await expect(first).resolves.toBe(true)
	})
})
