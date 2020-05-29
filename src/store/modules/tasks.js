import TaskService from '../../services/task'
import TaskAssigneeService from '../../services/taskAssignee'
import TaskAssigneeModel from '../../models/taskAssignee'
import LabelTaskModel from '../../models/labelTask'
import LabelTaskService from '../../services/labelTask'

export default {
	namespaced: true,
	state: () => ({}),
	actions: {
		update(ctx, task) {
			const taskService = new TaskService()
			return taskService.update(task)
				.then(t => {
					ctx.commit('kanban/setTaskInBucket', t, {root: true})
					return Promise.resolve(t)
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
		delete(ctx, task) {
			const taskService = new TaskService()
			return taskService.delete(task)
				.then(t => {
					ctx.commit('kanban/removeTaskInBucket', task, {root: true})
					return Promise.resolve(t)
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
		// Adds a task attachment in store.
		// This is an action to be able to commit other mutations
		addTaskAttachment(ctx, {taskId, attachment}) {
			const t = ctx.rootGetters['kanban/getTaskById'](taskId)
			if (t.task === null) {
				return
			}
			t.task.attachments.push(attachment)
			ctx.commit('kanban/setTaskInBucketByIndex', t, {root: true})
		},
		addAssignee(ctx, {user, taskId}) {

			const taskAssignee = new TaskAssigneeModel({userId: user.id, taskId: taskId})
			const taskAssigneeService = new TaskAssigneeService()

			return taskAssigneeService.create(taskAssignee)
				.then(r => {
					const t = ctx.rootGetters['kanban/getTaskById'](taskId)
					if (t.task === null) {
						// Don't try further adding a label if the task is not in kanban
						// Usually this means the kanban board hasn't been accessed until now.
						// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
						console.debug('Could not add assignee to task in kanban, task not found', t)
						return Promise.resolve(r)
					}
					t.task.assignees.push(user)
					ctx.commit('kanban/setTaskInBucketByIndex', t, {root: true})
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
		removeAssignee(ctx, {user, taskId}) {

			const taskAssignee = new TaskAssigneeModel({userId: user.id, taskId: taskId})
			const taskAssigneeService = new TaskAssigneeService()

			return taskAssigneeService.delete(taskAssignee)
				.then(r => {
					const t = ctx.rootGetters['kanban/getTaskById'](taskId)
					if (t.task === null) {
						// Don't try further adding a label if the task is not in kanban
						// Usually this means the kanban board hasn't been accessed until now.
						// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
						console.debug('Could not remove assignee from task in kanban, task not found', t)
						return Promise.resolve(r)
					}

					for (const a in t.task.assignees) {
						if (t.task.assignees[a].id === user.id) {
							t.task.assignees.splice(a, 1)
							break
						}
					}

					ctx.commit('kanban/setTaskInBucketByIndex', t, {root: true})
					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})

		},
		addLabel(ctx, {label, taskId}) {

			const labelTaskService = new LabelTaskService()
			const labelTask = new LabelTaskModel({taskId: taskId, labelId: label.id})

			return labelTaskService.create(labelTask)
				.then(r => {
					const t = ctx.rootGetters['kanban/getTaskById'](taskId)
					if (t.task === null) {
						// Don't try further adding a label if the task is not in kanban
						// Usually this means the kanban board hasn't been accessed until now.
						// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
						console.debug('Could not add label to task in kanban, task not found', t)
						return Promise.resolve(r)
					}
					t.task.labels.push(label)
					ctx.commit('kanban/setTaskInBucketByIndex', t, {root: true})

					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
		removeLabel(ctx, {label, taskId}) {

			const labelTaskService = new LabelTaskService()
			const labelTask = new LabelTaskModel({taskId: taskId, labelId: label.id})

			return labelTaskService.delete(labelTask)
				.then(r => {
					const t = ctx.rootGetters['kanban/getTaskById'](taskId)
					if (t.task === null) {
						// Don't try further adding a label if the task is not in kanban
						// Usually this means the kanban board hasn't been accessed until now.
						// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
						console.debug('Could not remove label from task in kanban, task not found', t)
						return Promise.resolve(r)
					}

					// Remove the label from the list
					for (const l in t.task.labels) {
						if (t.task.labels[l].id === label.id) {
							t.task.labels.splice(l, 1)
							break
						}
					}

					ctx.commit('kanban/setTaskInBucketByIndex', t, {root: true})

					return Promise.resolve(r)
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
	},
}