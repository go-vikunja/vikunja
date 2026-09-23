import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {
	createTaskDraft,
	createReminderDraft,
	getTaskIdentifier,
	getHexColor,
	parseRepeatAfter,
	repeatAfterToSeconds,
	mergeTask,
	mapTasksDeep,
	removeTask,
	moveTaskToBucket,
	getDefaultBucketId,
	buildDefaultRemindersForQuickAdd,
	buildQuickAddTask,
} from './task'
import {parseTaskText, PREFIXES, PrefixMode, type ParsedTaskText} from '@/modules/quickAddMagic'
import type {Bucket, Task} from '@/client/generated'
import type {IRepeatAfter} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {useAuthStore} from '@/stores/auth'
import UserSettingsModel from '@/models/userSettings'

describe('task domain helpers', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})
	it('fills wire defaults for fields explicitly set to undefined', () => {
		expect(createTaskDraft({title: ' New task ', due_date: '2026-09-17T12:00:00Z'})).toMatchObject({
			title: 'New task',
			due_date: '2026-09-17T12:00:00Z',
			repeat_after: 0,
		})
		expect(createTaskDraft({labels: undefined}).labels).toEqual([])
	})
	it('fills wire defaults for reminders', () => {
		expect(createReminderDraft()).toEqual({relative_period: 0, relative_to: '', reminder: ''})
		expect(createReminderDraft({relative_period: 60})).toEqual({
			relative_period: 60,
			relative_to: '',
			reminder: '',
		})
		expect(createReminderDraft({reminder: undefined})).toEqual({
			relative_period: 0,
			relative_to: '',
			reminder: '',
		})
	})
	it('formats identifiers and optional colors without changing task data', () => {
		expect(getTaskIdentifier({identifier: 'WORK-2', index: 2})).toBe('WORK-2')
		expect(getTaskIdentifier({identifier: '-2', index: 2})).toBe('#2')
		expect(getTaskIdentifier({index: 2})).toBe('#2')
		expect(getTaskIdentifier(null)).toBe('')
		expect(getHexColor('aabbcc')).toBe('#aabbcc')
		expect(getHexColor('#aabbcc')).toBe('#aabbcc')
		expect(getHexColor('#')).toBeUndefined()
	})
	it('converts repeat periods at the form boundary', () => {
		expect(parseRepeatAfter(604800)).toEqual({type: 'weeks', amount: 1})
		expect(repeatAfterToSeconds({type: 'days', amount: 3})).toBe(259200)
		expect(repeatAfterToSeconds({type: 'minutes', amount: 30})).toBe(1800)
		expect(repeatAfterToSeconds(3600)).toBe(3600)
		expect(repeatAfterToSeconds(undefined)).toBe(0)
		expect(repeatAfterToSeconds({type: 'days'} as IRepeatAfter)).toBe(0)
	})
	it('replaces nested tasks without losing view positions or expansions', () => {
		const tasks: Task[] = [{
			id: 1,
			title: 'old',
			position: 42,
			labels: [{id: 2}],
			related_tasks: {subtask: [{id: 2, title: 'old child'}]},
		}]
		expect(mapTasksDeep(tasks, 1, task => mergeTask(task, {id: 1, title: 'new', position: 0}))).toMatchObject([{
			title: 'new',
			position: 42,
			labels: [{id: 2}],
		}])
		expect(
			mapTasksDeep(tasks, 2, task => mergeTask(task, {id: 2, title: 'new child'}))[0].related_tasks?.subtask,
		).toEqual([{id: 2, title: 'new child'}])
		const cyclic: Task = {id: 1, related_tasks: {subtask: [{id: 2, related_tasks: {parenttask: [{id: 1}]}}]}}
		expect(mapTasksDeep([cyclic], 1, task => mergeTask(task, cyclic))).toEqual([cyclic])
		expect(removeTask(tasks, 2)[0].related_tasks?.subtask).toEqual([])
		expect(removeTask(tasks, 1)).toEqual([])
		expect(tasks[0].title).toBe('old')
	})
	it('keeps cached attachments, reactions and creator when a response omits them', () => {
		const cached: Task = {id: 1, attachments: [{id: 5}], reactions: {'👍': [{id: 6}]}, created_by: {id: 7}}
		expect(mergeTask(cached, {id: 1, title: 'new'})).toEqual({...cached, title: 'new'})
		expect(mergeTask(cached, {id: 1, attachments: [], reactions: {}, created_by: {id: 8}})).toEqual({
			id: 1,
			attachments: [],
			reactions: {},
			created_by: {id: 8},
		})
	})
	it('moves only between loaded buckets and preserves counts', () => {
		const buckets: Bucket[] = [{id: 1, count: 10, tasks: [{id: 4, bucket_id: 1}]}, {id: 2, count: 5, tasks: []}]
		expect(moveTaskToBucket(buckets, {id: 4}, 99)).toBe(buckets)
		expect(moveTaskToBucket(buckets, {id: 7}, 2)).toBe(buckets)
		expect(moveTaskToBucket(buckets, {id: 4}, 1)).toBe(buckets)
		const moved = moveTaskToBucket(buckets, {id: 4, title: 'moved'}, 2)
		expect(moved).toMatchObject([{count: 9, tasks: []}, {count: 6, tasks: [{id: 4, bucket_id: 2}]}])
		expect(buckets[0].count).toBe(10)
		expect(getDefaultBucketId({default_bucket_id: 2}, buckets)).toBe(2)
		expect(getDefaultBucketId({}, buckets)).toBe(1)
	})
	it('keeps cached view data and sibling relations when moving a task to another bucket', () => {
		const buckets: Bucket[] = [
			{
				id: 1,
				count: 2,
				tasks: [
					{id: 4, bucket_id: 1, position: 8, labels: [{id: 7}]},
					{id: 5, related_tasks: {subtask: [{id: 4, title: 'child'}]}},
				],
			},
			{id: 2, count: 0, tasks: []},
		]
		const moved = moveTaskToBucket(buckets, {id: 4, position: 0, labels: null}, 2)
		expect(moved[1].tasks).toEqual([{id: 4, bucket_id: 2, position: 8, labels: [{id: 7}]}])
		expect(moved[0].tasks).toEqual([{id: 5, related_tasks: {subtask: [{id: 4, title: 'child'}]}}])
	})
	it('builds relative default reminders only for tasks with a due date', () => {
		const defaults = [{relative_period: -900, relative_to: 'start_date'}]
		expect(buildDefaultRemindersForQuickAdd(defaults, '')).toEqual([])
		expect(buildDefaultRemindersForQuickAdd(defaults, '2026-09-17T12:00:00Z')).toEqual([{
			relative_period: -900,
			relative_to: 'due_date',
		}])
	})
	it('cleans only resolved assignees while preserving quick-add labels', () => {
		const input = {title: 'Task @alice @missing *label', project_id: 1}
		const parsed = parseTaskText(input.title, PrefixMode.Default)
		const task = buildQuickAddTask(
			parsed,
			input,
			PREFIXES[PrefixMode.Default],
			[{id: 2, username: 'alice', match: 'alice'}],
		)
		expect(task.title).toBe('Task @missing')
		expect(task.assignees).toEqual([{id: 2, username: 'alice'}])
		expect(parsed.labels).toEqual(['label'])
	})
	it('keeps the raw input title when quick add magic parses no text', () => {
		const parsed: ParsedTaskText = {
			text: '',
			date: null,
			labels: [],
			project: null,
			priority: null,
			assignees: [],
			repeats: null,
		}
		expect(
			buildQuickAddTask(parsed, {title: ' *label ', project_id: 1}, PREFIXES[PrefixMode.Default], []),
		).toMatchObject({
			title: '*label',
			project_id: 1,
			priority: 0,
			repeat_after: 0,
			repeat_mode: 0,
			labels: [],
		})
	})
	it('marks monthly repeats with the month repeat mode instead of an interval', () => {
		const parsed = parseTaskText('Task every month', PrefixMode.Default)
		const task = buildQuickAddTask(
			parsed,
			{title: 'Task every month', project_id: 1},
			PREFIXES[PrefixMode.Default],
			[],
		)
		expect(task.title).toBe('Task')
		expect(task.repeat_mode).toBe(TASK_REPEAT_MODES.REPEAT_MODE_MONTH)
		expect(task.repeat_after).toBe(0)
	})
	describe('repeating quick add without a date', () => {
		afterEach(() => {
			vi.useRealTimers()
		})
		it('sets the first due date to today at the default due time', () => {
			vi.useFakeTimers()
			vi.setSystemTime(new Date(2026, 8, 23, 9, 15))
			const settings = new UserSettingsModel()
			settings.frontendSettings.defaultDueTime = '14:30'
			useAuthStore().setUserSettings(settings)
			const defaults = [{relative_period: -900, relative_to: 'due_date'}]
			const parsed = parseTaskText('Call mom every day', PrefixMode.Default)
			const task = buildQuickAddTask(
				parsed,
				{title: 'Call mom every day', project_id: 1},
				PREFIXES[PrefixMode.Default],
				[],
				defaults,
			)
			expect(task.due_date).toBe(new Date(2026, 8, 23, 14, 30).toISOString())
			expect(task.reminders).toEqual(defaults)
		})
		it('keeps an explicitly parsed date', () => {
			vi.useFakeTimers()
			vi.setSystemTime(new Date(2026, 8, 23, 9, 15))
			const parsed = parseTaskText('Call mom every day at 11:42', PrefixMode.Default)
			const task = buildQuickAddTask(
				parsed,
				{title: 'Call mom every day at 11:42', project_id: 1},
				PREFIXES[PrefixMode.Default],
				[],
			)
			expect(task.due_date).toBe(new Date(2026, 8, 23, 11, 42).toISOString())
		})
	})
	it('retains embedded resources omitted from a task write response', () => {
		const task = {id: 1, labels: [{id: 2}], related_tasks: {subtask: [{id: 3}]}, assignees: [{id: 4}]}
		expect(
			mapTasksDeep([task], 1, current => mergeTask(current, {
				id: 1,
				title: 'new',
				labels: null,
				assignees: null,
			}))[0],
		).toMatchObject({...task, title: 'new'})
	})
})
