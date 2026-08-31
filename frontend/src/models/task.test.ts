import {describe, expect, it} from 'vitest'

import BucketModel from './bucket'
import ProjectModel from './project'
import TaskModel from './task'
import TaskDuplicateModel from './taskDuplicateModel'

const generatedLabel = {
	id: 1,
	title: 'Label',
	hex_color: 'ff006e',
	created_by: {id: 2, username: 'creator'},
}

function expectGeneratedLabel(label: Record<string, unknown>) {
	expect(label).toMatchObject(generatedLabel)
	expect(label).not.toHaveProperty('hexColor')
	expect(label).not.toHaveProperty('createdBy')
}

describe('TaskModel labels', () => {
	it('preserves generated label field casing', () => {
		const task = new TaskModel({
			labels: [generatedLabel],
		})

		expectGeneratedLabel(task.labels[0])
	})

	it.each([
		['bucket', () => new BucketModel({tasks: [{labels: [generatedLabel]}], created_by: {}} as never).tasks[0]],
		['project', () => new ProjectModel({tasks: [{labels: [generatedLabel]}], owner: {}, views: []} as never).tasks[0]],
		['duplicate', () => new TaskDuplicateModel({duplicated_task: {labels: [generatedLabel]}} as never).duplicatedTask],
	])('restores generated casing after %s model conversion', (_, createTask) => {
		const task = createTask()

		expectGeneratedLabel(task!.labels[0])
	})

	it('restores generated casing on related tasks', () => {
		const task = new TaskModel({
			related_tasks: {
				subtask: [{labels: [generatedLabel]}],
			},
		} as never)

		expectGeneratedLabel(task.relatedTasks.subtask![0].labels[0])
	})
})
