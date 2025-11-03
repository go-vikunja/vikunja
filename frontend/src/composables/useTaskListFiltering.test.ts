import {describe, it, expect} from 'vitest'
import {shouldShowTaskInListView} from './useTaskListFiltering'
import type {ITask} from '@/modelTypes/ITask'

describe('shouldShowTaskInListView', () => {
	it('should hide subtasks when parent is in the same project', () => {
		const parentTask: Partial<ITask> = {
			id: 1,
			title: 'Parent Task',
			projectId: 100,
			relatedTasks: {},
		}

		const subtask: Partial<ITask> = {
			id: 2,
			title: 'Subtask',
			projectId: 100,
			relatedTasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task',
					projectId: 100,
				} as ITask],
			},
		}

		const allTasks = [parentTask, subtask] as ITask[]

		expect(shouldShowTaskInListView(parentTask as ITask, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(subtask as ITask, allTasks)).toBe(false)
	})

	it('should show subtasks when parent is in a different project', () => {
		const parentTask: Partial<ITask> = {
			id: 1,
			title: 'Parent Task in Project A',
			projectId: 100,
		}

		const subtask: Partial<ITask> = {
			id: 2,
			title: 'Subtask in Project B',
			projectId: 200,
			relatedTasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task in Project A',
					projectId: 100,
				} as ITask],
			},
		}

		// In Project B's view, we only see the subtask
		const tasksInProjectB = [subtask] as ITask[]

		expect(shouldShowTaskInListView(subtask as ITask, tasksInProjectB)).toBe(true)
	})

	it('should show tasks with no parents', () => {
		const task: Partial<ITask> = {
			id: 1,
			title: 'Regular Task',
			projectId: 100,
			relatedTasks: {},
		}

		const allTasks = [task] as ITask[]

		expect(shouldShowTaskInListView(task as ITask, allTasks)).toBe(true)
	})

	it('should show tasks with undefined relatedTasks', () => {
		const task: Partial<ITask> = {
			id: 1,
			title: 'Regular Task',
			projectId: 100,
		}

		const allTasks = [task] as ITask[]

		expect(shouldShowTaskInListView(task as ITask, allTasks)).toBe(true)
	})

	it('should show tasks with empty parenttask array', () => {
		const task: Partial<ITask> = {
			id: 1,
			title: 'Regular Task',
			projectId: 100,
			relatedTasks: {
				parenttask: [],
			},
		}

		const allTasks = [task] as ITask[]

		expect(shouldShowTaskInListView(task as ITask, allTasks)).toBe(true)
	})

	it('should handle multiple levels of nesting within same project', () => {
		const grandparent: Partial<ITask> = {
			id: 1,
			title: 'Grandparent',
			projectId: 100,
			relatedTasks: {},
		}

		const parent: Partial<ITask> = {
			id: 2,
			title: 'Parent',
			projectId: 100,
			relatedTasks: {
				parenttask: [{id: 1, title: 'Grandparent', projectId: 100} as ITask],
			},
		}

		const child: Partial<ITask> = {
			id: 3,
			title: 'Child',
			projectId: 100,
			relatedTasks: {
				parenttask: [{id: 2, title: 'Parent', projectId: 100} as ITask],
			},
		}

		const allTasks = [grandparent, parent, child] as ITask[]

		expect(shouldShowTaskInListView(grandparent as ITask, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(parent as ITask, allTasks)).toBe(false)
		expect(shouldShowTaskInListView(child as ITask, allTasks)).toBe(false)
	})

	it('should show task if it has multiple parents and none are in view', () => {
		const subtask: Partial<ITask> = {
			id: 3,
			title: 'Subtask with multiple parents',
			projectId: 300,
			relatedTasks: {
				parenttask: [
					{id: 1, title: 'Parent 1', projectId: 100} as ITask,
					{id: 2, title: 'Parent 2', projectId: 200} as ITask,
				],
			},
		}

		// In Project 300's view, neither parent is present
		const tasksInProject300 = [subtask] as ITask[]

		expect(shouldShowTaskInListView(subtask as ITask, tasksInProject300)).toBe(true)
	})

	it('should hide task if it has multiple parents and at least one is in view', () => {
		const parent1: Partial<ITask> = {
			id: 1,
			title: 'Parent 1',
			projectId: 100,
		}

		const parent2: Partial<ITask> = {
			id: 2,
			title: 'Parent 2',
			projectId: 100,
		}

		const subtask: Partial<ITask> = {
			id: 3,
			title: 'Subtask with multiple parents',
			projectId: 100,
			relatedTasks: {
				parenttask: [
					{id: 1, title: 'Parent 1', projectId: 100} as ITask,
					{id: 2, title: 'Parent 2', projectId: 100} as ITask,
				],
			},
		}

		const allTasks = [parent1, parent2, subtask] as ITask[]

		expect(shouldShowTaskInListView(subtask as ITask, allTasks)).toBe(false)
	})
})
