import {describe, expect, it} from 'vitest'
import {createTaskDraft, getTaskIdentifier, getHexColor, parseRepeatAfter, repeatAfterToSeconds, replaceTask, removeTask, moveTaskToBucket, getDefaultBucketId, buildDefaultRemindersForQuickAdd, buildQuickAddTask} from './task'
import {parseTaskText, PrefixMode} from '@/modules/quickAddMagic'
import type {Bucket, Task} from '@/client/generated'
import type {IRepeatAfter} from '@/types/IRepeatAfter'

describe('task domain helpers', () => {
	it('provides independent drafts using wire fields and timestamp strings', () => {
		const first = createTaskDraft({title: ' New task ', due_date: '2026-09-17T12:00:00Z'})
		first.reminders!.push({relative_period: -60})
		expect(createTaskDraft().reminders).toEqual([])
		expect(first).toMatchObject({title: 'New task', due_date: '2026-09-17T12:00:00Z', repeat_after: 0})
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
		const tasks: Task[] = [{id: 1, title: 'old', position: 42, labels: [{id: 2}], related_tasks: {subtask: [{id: 2, title: 'old child'}]}}]
		expect(replaceTask(tasks, {id: 1, title: 'new', position: 0})).toMatchObject([{title: 'new', position: 42, labels: [{id: 2}]}])
		expect(replaceTask(tasks, {id: 2, title: 'new child'})[0].related_tasks?.subtask).toEqual([{id: 2, title: 'new child'}])
		expect(removeTask(tasks, 2)[0].related_tasks?.subtask).toEqual([])
		expect(tasks[0].title).toBe('old')
	})
	it('moves only between loaded buckets and preserves counts', () => {
		const buckets: Bucket[] = [{id: 1, count: 10, tasks: [{id: 4, bucket_id: 1}]}, {id: 2, count: 5, tasks: []}]
		expect(moveTaskToBucket(buckets, {id: 4}, 99)).toBe(buckets)
		const moved = moveTaskToBucket(buckets, {id: 4, title: 'moved'}, 2)
		expect(moved).toMatchObject([{count: 9, tasks: []}, {count: 6, tasks: [{id: 4, bucket_id: 2}]}])
		expect(buckets[0].count).toBe(10)
		expect(getDefaultBucketId({default_bucket_id: 2}, buckets)).toBe(2)
		expect(getDefaultBucketId({}, buckets)).toBe(1)
	})
	it('keeps cached view data and sibling relations when moving a task to another bucket', () => {
		const buckets: Bucket[] = [
			{id: 1, count: 2, tasks: [{id: 4, bucket_id: 1, position: 8, labels: [{id: 7}]}, {id: 5, related_tasks: {subtask: [{id: 4, title: 'child'}]}}]},
			{id: 2, count: 0, tasks: []},
		]
		const moved = moveTaskToBucket(buckets, {id: 4, position: 0, labels: null}, 2)
		expect(moved[1].tasks).toEqual([{id: 4, bucket_id: 2, position: 8, labels: [{id: 7}]}])
		expect(moved[0].tasks).toEqual([{id: 5, related_tasks: {subtask: [{id: 4, title: 'child'}]}}])
	})
	it('builds relative default reminders only for tasks with a due date', () => {
		const defaults = [{relative_period: -900, relative_to: 'start_date'}]
		expect(buildDefaultRemindersForQuickAdd(defaults, '')).toEqual([])
		expect(buildDefaultRemindersForQuickAdd(defaults, '2026-09-17T12:00:00Z')).toEqual([{relative_period: -900, relative_to: 'due_date'}])
	})
	it('cleans only resolved assignees while preserving quick-add labels', () => {
		const input = {title: 'Task @alice @missing *label', project_id: 1}
		const parsed = parseTaskText(input.title, PrefixMode.Default)
		const task = buildQuickAddTask(parsed, input, PrefixMode.Default, [{id: 2, username: 'alice', match: 'alice'}])
		expect(task.title).toBe('Task @missing')
		expect(task.assignees).toEqual([{id: 2, username: 'alice'}])
		expect(parsed.labels).toEqual(['label'])
	})
})
