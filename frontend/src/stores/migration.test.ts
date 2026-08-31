import {setActivePinia, createPinia} from 'pinia'
import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'

import {useMigrationStore} from './migration'
import type {MigrationStatus} from '@/services/migrator/abstractMigration'

const {refreshProjectsMock} = vi.hoisted(() => ({
	refreshProjectsMock: vi.fn(),
}))

vi.mock('@/client/queries/projects', () => ({
	refreshProjects: refreshProjectsMock,
}))

const POLL_INTERVAL = 3000
const NEVER = '0001-01-01T00:00:00Z'

function status(overrides: Partial<MigrationStatus> = {}): MigrationStatus {
	return {
		started_at: '2024-01-15T10:00:00Z',
		finished_at: NEVER,
		error_kind: '',
		error_message: '',
		...overrides,
	}
}

// Lets the poll's awaited promises settle between timer ticks.
async function tick() {
	await vi.advanceTimersByTimeAsync(POLL_INTERVAL)
}

describe('migration store', () => {
	beforeEach(() => {
		vi.useFakeTimers()
		setActivePinia(createPinia())
		refreshProjectsMock.mockReset()
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('does not poll until a view starts it', async () => {
		const getStatus = vi.fn().mockResolvedValue(status())
		useMigrationStore()

		await tick()

		expect(getStatus).not.toHaveBeenCalled()
	})

	it('keeps polling while the migration runs', async () => {
		const getStatus = vi.fn().mockResolvedValue(status())
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()
		await tick()

		expect(getStatus).toHaveBeenCalledTimes(2)
		expect(store.isFinished).toBe(false)
	})

	it('finishes and reloads the projects on success', async () => {
		const getStatus = vi.fn().mockResolvedValue(status({finished_at: '2024-01-15T10:05:00Z'}))
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()
		await tick()

		expect(store.isFinished).toBe(true)
		expect(store.hasFailed).toBe(false)
		expect(refreshProjectsMock).toHaveBeenCalledTimes(1)
		expect(getStatus).toHaveBeenCalledTimes(1)
	})

	it('does not reload the projects when the migration failed', async () => {
		const getStatus = vi.fn().mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'interrupted',
		}))
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()

		expect(store.hasFailed).toBe(true)
		expect(store.failureKey).toBe('migrate.failure.interrupted')
		expect(refreshProjectsMock).not.toHaveBeenCalled()
	})

	it('renders a detail failure with the raw message', async () => {
		const getStatus = vi.fn().mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'detail',
			error_message: 'row 4 has no title',
		}))
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()

		expect(store.failureKey).toBe('migrate.migrationFailed')
		expect(store.errorMessage).toBe('row 4 has no title')
	})

	it('falls back to the generic text for a kind it does not know', async () => {
		const getStatus = vi.fn().mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'something-new' as never,
		}))
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()

		expect(store.failureKey).toBe('migrate.failure.reported')
	})

	it('stop cancels the poller', async () => {
		const getStatus = vi.fn().mockResolvedValue(status())
		const store = useMigrationStore()
		store.start({getStatus})
		store.stop()

		await tick()
		await tick()

		expect(getStatus).not.toHaveBeenCalled()
	})

	it('ignores the response of a poll started before stop', async () => {
		let resolveStatus: (s: MigrationStatus) => void = () => {}
		const getStatus = vi.fn().mockReturnValue(new Promise<MigrationStatus>(resolve => {
			resolveStatus = resolve
		}))
		const store = useMigrationStore()
		store.start({getStatus})

		await tick()
		store.stop()
		resolveStatus(status({finished_at: '2024-01-15T10:05:00Z'}))
		await tick()

		expect(store.isFinished).toBe(false)
		expect(refreshProjectsMock).not.toHaveBeenCalled()
	})

	it('gives up after the consecutive failure cap', async () => {
		const getStatus = vi.fn().mockRejectedValue(new Error('nope'))
		const store = useMigrationStore()
		store.start({getStatus})

		for (let i = 0; i < 10; i++) {
			await tick()
		}

		expect(getStatus).toHaveBeenCalledTimes(5)
		expect(store.isFinished).toBe(false)
	})

	it('gives up once the deadline passed', async () => {
		const getStatus = vi.fn().mockResolvedValue(status())
		const store = useMigrationStore()
		store.start({getStatus})

		await vi.advanceTimersByTimeAsync(20 * 60 * 1000)
		const callsAtDeadline = getStatus.mock.calls.length
		await tick()
		await tick()

		expect(getStatus).toHaveBeenCalledTimes(callsAtDeadline)
		expect(store.isFinished).toBe(false)
	})

	it('start resets the state of a previous migration', async () => {
		const failed = vi.fn().mockResolvedValue(status({
			finished_at: '2024-01-15T10:05:00Z',
			error_kind: 'queue',
		}))
		const store = useMigrationStore()
		store.start({getStatus: failed})
		await tick()

		store.start({getStatus: vi.fn().mockResolvedValue(status())})

		expect(store.isFinished).toBe(false)
		expect(store.hasFailed).toBe(false)
	})
})
