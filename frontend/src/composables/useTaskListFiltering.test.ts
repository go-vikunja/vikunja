import {describe, it, expect} from 'vitest'
import {shouldShowTaskInListView} from './useTaskListFiltering'
import {normalizeTask} from '@/client/queries/tasks'

describe('shouldShowTaskInListView', () => {
	it('should hide subtasks when parent is in the same project', () => {
		const parentTask = normalizeTask({
			id: 1,
			title: 'Parent Task',
			project_id: 100,
			related_tasks: {},
		})

		const subtask = normalizeTask({
			id: 2,
			title: 'Subtask',
			project_id: 100,
			related_tasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task',
					project_id: 100,
				}],
			},
		})

		const allTasks = [parentTask, subtask]

		expect(shouldShowTaskInListView(parentTask, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(subtask, allTasks)).toBe(false)
	})

	it('should show subtasks when parent is in a different project', () => {
		const parentTask = normalizeTask({
			id: 1,
			title: 'Parent Task in Project A',
			project_id: 100,
		})

		const subtask = normalizeTask({
			id: 2,
			title: 'Subtask in Project B',
			project_id: 200,
			related_tasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task in Project A',
					project_id: 100,
				}],
			},
		})

		// In Project B's view, we only see the subtask
		const tasksInProjectB = [subtask]

		expect(shouldShowTaskInListView(subtask, tasksInProjectB)).toBe(true)
	})

	it.each([
		['no related tasks', {}],
		['undefined related tasks', undefined],
		['an empty parenttask array', {parenttask: []}],
	])('should show a task with %s', (_case, related_tasks) => {
		const task = normalizeTask({
			id: 1,
			title: 'Regular Task',
			project_id: 100,
			related_tasks,
		})

		expect(shouldShowTaskInListView(task, [task])).toBe(true)
	})

	it('should handle multiple levels of nesting within same project', () => {
		const grandparent = normalizeTask({
			id: 1,
			title: 'Grandparent',
			project_id: 100,
			related_tasks: {},
		})

		const parent = normalizeTask({
			id: 2,
			title: 'Parent',
			project_id: 100,
			related_tasks: {
				parenttask: [{id: 1, title: 'Grandparent', project_id: 100}],
			},
		})

		const child = normalizeTask({
			id: 3,
			title: 'Child',
			project_id: 100,
			related_tasks: {
				parenttask: [{id: 2, title: 'Parent', project_id: 100}],
			},
		})

		const allTasks = [grandparent, parent, child]

		expect(shouldShowTaskInListView(grandparent, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(parent, allTasks)).toBe(false)
		expect(shouldShowTaskInListView(child, allTasks)).toBe(false)
	})

	it('should show task if it has multiple parents and none are in view', () => {
		const subtask = normalizeTask({
			id: 3,
			title: 'Subtask with multiple parents',
			project_id: 300,
			related_tasks: {
				parenttask: [
					{id: 1, title: 'Parent 1', project_id: 100},
					{id: 2, title: 'Parent 2', project_id: 200},
				],
			},
		})

		// In Project 300's view, neither parent is present
		const tasksInProject300 = [subtask]

		expect(shouldShowTaskInListView(subtask, tasksInProject300)).toBe(true)
	})

	it('should hide task if it has multiple parents and at least one is in view', () => {
		const parent1 = normalizeTask({
			id: 1,
			title: 'Parent 1',
			project_id: 100,
		})

		const parent2 = normalizeTask({
			id: 2,
			title: 'Parent 2',
			project_id: 100,
		})

		const subtask = normalizeTask({
			id: 3,
			title: 'Subtask with multiple parents',
			project_id: 100,
			related_tasks: {
				parenttask: [
					{id: 1, title: 'Parent 1', project_id: 100},
					{id: 2, title: 'Parent 2', project_id: 100},
				],
			},
		})

		const allTasks = [parent1, parent2, subtask]

		expect(shouldShowTaskInListView(subtask, allTasks)).toBe(false)
	})

})
