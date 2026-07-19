import {setActivePinia, createPinia} from 'pinia'
import {beforeEach, describe, expect, it, vi} from 'vitest'

vi.mock('@/router', () => ({
	default: {
		currentRoute: {value: {params: {}}},
		isReady: () => Promise.resolve(),
	},
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({t: (key: string) => key}),
	createI18n: () => ({global: {t: (key: string) => key}}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({setHasTasks: vi.fn()}),
}))

import {buildDefaultRemindersForQuickAdd, runWrites, useTaskStore} from './tasks'
import {useLabelStore} from './labels'
import LabelModel from '@/models/label'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import type {ILabel} from '@/modelTypes/ILabel'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

const aDefault: ITaskReminder = {
	reminder: null,
	relativePeriod: -3600,
	relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
} as ITaskReminder

describe('buildDefaultRemindersForQuickAdd', () => {
	it('returns empty array when due date is null', () => {
		expect(buildDefaultRemindersForQuickAdd([aDefault], null)).toEqual([])
	})

	it('returns empty array when defaults are undefined', () => {
		expect(buildDefaultRemindersForQuickAdd(undefined, '2026-05-01T00:00:00.000Z')).toEqual([])
	})

	it('returns empty array when defaults are empty', () => {
		expect(buildDefaultRemindersForQuickAdd([], '2026-05-01T00:00:00.000Z')).toEqual([])
	})

	it('clones defaults with relativeTo locked to due_date', () => {
		const result = buildDefaultRemindersForQuickAdd([aDefault], '2026-05-01T00:00:00.000Z')
		expect(result).toHaveLength(1)
		expect(result[0].relativePeriod).toBe(-3600)
		expect(result[0].relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
		expect(result[0].reminder).toBeNull()
	})

	it('does not share references with the input array', () => {
		const defaults = [aDefault]
		const result = buildDefaultRemindersForQuickAdd(defaults, '2026-05-01T00:00:00.000Z')
		expect(result[0]).not.toBe(defaults[0])
	})

	it('forces relativeTo to due_date even if a default somehow had another value', () => {
		const weird = {...aDefault, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE} as ITaskReminder
		const result = buildDefaultRemindersForQuickAdd([weird], '2026-05-01T00:00:00.000Z')
		expect(result[0].relativeTo).toBe(REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE)
	})
})

describe('runWrites', () => {
	function deferredWrite() {
		const inFlight: string[] = []
		let maxConcurrent = 0
		const completed: string[] = []
		const write = async (item: string) => {
			inFlight.push(item)
			maxConcurrent = Math.max(maxConcurrent, inFlight.length)
			await Promise.resolve()
			inFlight.splice(inFlight.indexOf(item), 1)
			completed.push(item)
		}
		return {write, completed, getMaxConcurrent: () => maxConcurrent}
	}

	it('runs all writes in parallel when concurrent', async () => {
		const {write, completed, getMaxConcurrent} = deferredWrite()
		await runWrites(['a', 'b', 'c'], write, true)
		expect(completed).toHaveLength(3)
		expect(getMaxConcurrent()).toBeGreaterThan(1)
	})

	it('runs writes one at a time when not concurrent', async () => {
		const {write, completed, getMaxConcurrent} = deferredWrite()
		await runWrites(['a', 'b', 'c'], write, false)
		expect(completed).toEqual(['a', 'b', 'c'])
		expect(getMaxConcurrent()).toBe(1)
	})

	it('does nothing for an empty list', async () => {
		const {write, completed} = deferredWrite()
		await runWrites([], write, false)
		expect(completed).toHaveLength(0)
	})
})

describe('ensureLabelsExist', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('skips labels that fail to create and returns the resolved ones', async () => {
		const taskStore = useTaskStore()
		const labelStore = useLabelStore()
		labelStore.setLabels([{id: 1, title: 'existing'}] as ILabel[])

		vi.spyOn(labelStore, 'createLabel').mockImplementation(async label => {
			if (label.title === 'forbidden') {
				throw new Error('403')
			}
			return new LabelModel({id: 99, title: label.title})
		})

		const result = await taskStore.ensureLabelsExist(['existing', 'created', 'forbidden'])
		const titles = result.map(l => l.title)

		expect(titles).toContain('existing')
		expect(titles).toContain('created')
		expect(titles).not.toContain('forbidden')
		expect(result).toHaveLength(2)
	})
})
