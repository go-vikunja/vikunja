import {setActivePinia, createPinia} from 'pinia'
import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'

import {useMigrationStore} from './migration'
import type {Status as MigrationStatus} from '@/client/generated'
import {removeToken} from '@/helpers/auth'
import {queryClient} from '@/client/queryClient'
import {projectKeys} from '@/client/queries/projects'

const {getStatus} = vi.hoisted(() => ({getStatus: vi.fn()}))
vi.mock('@/client/generated', async importOriginal => ({...await importOriginal<typeof import('@/client/generated')>(), migrationCsvStatus: getStatus}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))

const POLL_INTERVAL = 3000
const NEVER = '0001-01-01T00:00:00Z'

function status(overrides: Partial<MigrationStatus> = {}) {
	return {data: {
		started_at: '2024-01-15T10:00:00Z',
		finished_at: NEVER,
		error_kind: '',
		error_message: '',
		...overrides,
	}}
}

// Lets the poll's awaited promises settle between timer ticks.
async function tick() {
	await vi.advanceTimersByTimeAsync(POLL_INTERVAL)
}

describe('migration store', () => {
	beforeEach(() => {
		vi.useFakeTimers()
		setActivePinia(createPinia())
		getStatus.mockReset()
		queryClient.clear()
		queryClient.setQueryData(projectKeys.list(), {projects: []})
	})

	afterEach(() => {
		useMigrationStore().stop()
		queryClient.clear()
		vi.useRealTimers()
	})

	it('does not poll until a view starts it', async () => {
		getStatus.mockResolvedValue(status())
		useMigrationStore()

		await tick()

		expect(getStatus).not.toHaveBeenCalled()
	})

	it('keeps polling while the migration runs', async () => {
		getStatus.mockResolvedValue(status())
		const store = useMigrationStore()
		store.start('csv')

		await tick()
		await tick()

		expect(getStatus).toHaveBeenCalledTimes(2)
		expect(store.isFinished).toBe(false)
	})

	it('finishes and reloads the projects on success', async () => {
		getStatus.mockResolvedValue(status({finished_at: '2024-01-15T10:05:00Z'}))
		const store = useMigrationStore()
		store.start('csv')

		await tick()
		await tick()

		expect(store.isFinished).toBe(true)
		expect(store.hasFailed).toBe(false)
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(true)
		expect(getStatus).toHaveBeenCalledTimes(1)
	})

	it('does not reload the projects when the migration failed', async () => {
		getStatus.mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'interrupted',
		}))
		const store = useMigrationStore()
		store.start('csv')

		await tick()

		expect(store.hasFailed).toBe(true)
		expect(store.failureKey).toBe('migrate.failure.interrupted')
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(false)
	})

	it('renders a detail failure with the raw message', async () => {
		getStatus.mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'detail',
			error_message: 'row 4 has no title',
		}))
		const store = useMigrationStore()
		store.start('csv')

		await tick()

		expect(store.failureKey).toBe('migrate.migrationFailed')
		expect(store.errorMessage).toBe('row 4 has no title')
	})

	it('falls back to the generic text for a kind it does not know', async () => {
		getStatus.mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'something-new' as never,
		}))
		const store = useMigrationStore()
		store.start('csv')

		await tick()

		expect(store.failureKey).toBe('migrate.failure.reported')
	})

	it('stop cancels the poller', async () => {
		getStatus.mockResolvedValue(status())
		const store = useMigrationStore()
		store.start('csv')
		store.stop()

		await tick()
		await tick()

		expect(getStatus).not.toHaveBeenCalled()
	})

	it('ignores the response of a poll started before stop', async () => {
		let resolveStatus: (s: ReturnType<typeof status>) => void = () => {}
		getStatus.mockReturnValue(new Promise<ReturnType<typeof status>>(resolve => {
			resolveStatus = resolve
		}))
		const store = useMigrationStore()
		store.start('csv')

		await tick()
		store.stop()
		resolveStatus(status({finished_at: '2024-01-15T10:05:00Z'}))
		await tick()

		expect(store.isFinished).toBe(false)
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(false)
	})

	it('gives up after the consecutive failure cap', async () => {
		getStatus.mockRejectedValue(new Error('nope'))
		const store = useMigrationStore()
		store.start('csv')

		for (let i = 0; i < 10; i++) {
			await tick()
		}

		expect(getStatus).toHaveBeenCalledTimes(5)
		expect(store.isFinished).toBe(false)
	})

	it('gives up once the deadline passed', async () => {
		getStatus.mockResolvedValue(status())
		const store = useMigrationStore()
		store.start('csv')

		await vi.advanceTimersByTimeAsync(20 * 60 * 1000)
		const callsAtDeadline = getStatus.mock.calls.length
		await tick()
		await tick()

		expect(getStatus).toHaveBeenCalledTimes(callsAtDeadline)
		expect(store.isFinished).toBe(false)
	})

	it('start resets the state of a previous migration', async () => {
		getStatus.mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'queue',
		}))
		const store = useMigrationStore()
		store.start('csv')
		await tick()

		getStatus.mockResolvedValue(status())
		store.start('csv')

		expect(store.isFinished).toBe(false)
		expect(store.hasFailed).toBe(false)
	})
	it('stops polling after the account session changes', async () => {
		getStatus.mockResolvedValue(status())
		const store = useMigrationStore()
		store.start('csv')
		await tick()
		removeToken()
		await tick()
		expect(getStatus).toHaveBeenCalledTimes(1)
		expect(store.isFinished).toBe(false)
	})

})
