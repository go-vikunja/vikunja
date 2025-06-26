import {computed, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import router from '@/router'

import TaskService from '@/services/task'
import TaskAssigneeService from '@/services/taskAssignee'
import LabelTaskService from '@/services/labelTask'

import {cleanupItemText, parseTaskText, PREFIXES} from '@/modules/parseTaskText'

import TaskAssigneeModel from '@/models/taskAssignee'
import LabelTaskModel from '@/models/labelTask'
import LabelTask from '@/models/labelTask'
import TaskModel from '@/models/task'
import LabelModel from '@/models/label'

import type {ILabel} from '@/modelTypes/ILabel'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {IProject} from '@/modelTypes/IProject'

import {setModuleLoading} from '@/stores/helper'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {useAttachmentStore} from '@/stores/attachments'
import {useKanbanStore} from '@/stores/kanban'
import {useBaseStore} from '@/stores/base'
import ProjectUserService from '@/services/projectUsers'
import {useAuthStore} from '@/stores/auth'
import TaskCollectionService, {type TaskFilterParams} from '@/services/taskCollection'
import {getRandomColorHex} from '@/helpers/color/randomColor'
import {REPEAT_TYPES} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'

interface MatchedAssignee extends IUser {
	match: string,
}

// IDEA: maybe use a small fuzzy search here to prevent errors
function findPropertyByValue(object, key, value, fuzzy = false) {
	return Object.values(object).find(l => {
		if (fuzzy) {
			return l[key]?.toLowerCase().includes(value.toLowerCase())
		}
	
		return l[key]?.toLowerCase() === value.toLowerCase()
	})
}

// Check if the user exists in the search results
function validateUser(
	users: IUser[],
	query: IUser['username'] | IUser['name'] | IUser['email'],
) {
	if (users.length === 1) {
		return (
			findPropertyByValue(users, 'username', query, true) ||
			findPropertyByValue(users, 'name', query, true) ||
			findPropertyByValue(users, 'email', query, true)
		)
	}
	
	return (
		findPropertyByValue(users, 'username', query) ||
		findPropertyByValue(users, 'name', query) ||
		findPropertyByValue(users, 'email', query)
	)
}

// Check if the label exists
function validateLabel(labels: ILabel[], label: string) {
	return findPropertyByValue(labels, 'title', label)
}

async function addLabelToTask(task: ITask, label: ILabel) {
	const labelTask = new LabelTask({
		taskId: task.id,
		labelId: label.id,
	})
	const labelTaskService = new LabelTaskService()
	const response = await labelTaskService.create(labelTask)
	task.labels.push(label)
	return response
}

async function findAssignees(parsedTaskAssignees: string[], projectId: number): Promise<MatchedAssignee[]> {
	if (parsedTaskAssignees.length <= 0) {
		return []
	}

	const userService = new ProjectUserService()
	const assignees = parsedTaskAssignees.map(async a => {
		const users = (await userService.getAll({projectId}, {s: a}))
			.map(u => ({
				...u,
				match: a,
			}))
		return validateUser(users, a)
	})

	const validatedUsers = await Promise.all(assignees) 
	return validatedUsers.filter((item) => Boolean(item))
}

export const useTaskStore = defineStore('task', () => {
	const baseStore = useBaseStore()
	const kanbanStore = useKanbanStore()
	const attachmentStore = useAttachmentStore()
	const labelStore = useLabelStore()
	const projectStore = useProjectStore()
	const authStore = useAuthStore()

	const tasks = ref<{ [id: ITask['id']]: ITask }>({}) // TODO: or is this ITask[]
	const isLoading = ref(false)

	const hasTasks = computed(() => Object.keys(tasks.value).length > 0)

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setTasks(newTasks: ITask[]) {
		newTasks.forEach(task => {
			tasks.value[task.id] = task
		})
	}

	async function loadTasks(
		params: TaskFilterParams, 
		projectId: IProject['id'] | null = null,
	) {
		
		if (!params.filter_timezone || params.filter_timezone === '') {
			params.filter_timezone = authStore.settings.timezone
		}

		const cancel = setModuleLoading(setIsLoading)
		try {
			const model = {}
			let taskCollectionService = new TaskService()
			if (projectId !== null) {
				model.projectId = projectId
				taskCollectionService = new TaskCollectionService()
			}
			tasks.value = await taskCollectionService.getAll(model, params)
			baseStore.setHasTasks(tasks.value.length > 0)
			return tasks.value
		} finally {
			cancel()
		}
	}

	async function update(task: ITask) {
		const cancel = setModuleLoading(setIsLoading)

		const taskService = new TaskService()
		try {
			const updatedTask = await taskService.update(task)
			kanbanStore.ensureTaskIsInCorrectBucket(updatedTask)
			return updatedTask
		} finally {
			cancel()
		}
	}

	async function deleteTask(task: ITask) {
		const taskService = new TaskService()
		const response = await taskService.delete(task)
		kanbanStore.removeTaskInBucket(task)
		return response
	}

	// Adds a task attachment in store.
	// This is an action to be able to commit other mutations
	function addTaskAttachment({
		taskId,
		attachment,
	}: {
		taskId: ITask['id']
		attachment: IAttachment
	}) {
		const t = kanbanStore.getTaskById(taskId)
		if (t.task !== null) {
			const attachments = [
				...t.task.attachments,
				attachment,
			]

			const newTask = {
				...t,
				task: {
					...t.task,
					attachments,
				},
			}
			kanbanStore.setTaskInBucketByIndex(newTask)
		}
		attachmentStore.add(attachment)
	}

	async function addAssignee({
		user,
		taskId,
	}: {
		user: IUser,
		taskId: ITask['id']
	}) {
		const cancel = setModuleLoading(setIsLoading)
		
		try {
			const taskAssigneeService = new TaskAssigneeService()
			const r = await taskAssigneeService.create(new TaskAssigneeModel({
				userId: user.id,
				taskId: taskId,
			}))
			const t = kanbanStore.getTaskById(taskId)
			if (t.task === null) {
				// Don't try further adding a label if the task is not in kanban
				// Usually this means the kanban board hasn't been accessed until now.
				// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
				console.debug('Could not add assignee to task in kanban, task not found', t)
				return r
			}

			kanbanStore.setTaskInBucketByIndex({
				...t,
				task: {
					...t.task,
					assignees: [
						...t.task.assignees,
						user,
					],
				},
			})

			return r
		} finally {
			cancel()
		}
	}

	async function removeAssignee({
		user,
		taskId,
	}: {
		user: IUser,
		taskId: ITask['id']
	}) {
		const taskAssigneeService = new TaskAssigneeService()
		const response = await taskAssigneeService.delete(new TaskAssigneeModel({
			userId: user.id,
			taskId: taskId,
		}))
		const t = kanbanStore.getTaskById(taskId)
		if (t.task === null) {
			// Don't try further adding a label if the task is not in kanban
			// Usually this means the kanban board hasn't been accessed until now.
			// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
			console.debug('Could not remove assignee from task in kanban, task not found', t)
			return response
		}

		const assignees = t.task.assignees.filter(({ id }) => id !== user.id)

		kanbanStore.setTaskInBucketByIndex({
			...t,
			task: {
				...t.task,
				assignees,
			},
		})
		return response

	}

	async function addLabel({
		label,
		taskId,
	} : {
		label: ILabel,
		taskId: ITask['id']
	}) {
		const labelTaskService = new LabelTaskService()
		const r = await labelTaskService.create(new LabelTaskModel({
			taskId,
			labelId: label.id,
		}))
		const t = kanbanStore.getTaskById(taskId)
		if (t.task === null) {
			// Don't try further adding a label if the task is not in kanban
			// Usually this means the kanban board hasn't been accessed until now.
			// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
			console.debug('Could not add label to task in kanban, task not found', {taskId, t})
			return r
		}

		kanbanStore.setTaskInBucketByIndex({
			...t,
			task: {
				...t.task,
				labels: [
					...t.task.labels,
					label,
				],
			},
		})

		return r
	}

	async function removeLabel(
		{label, taskId}:
		{label: ILabel, taskId: ITask['id']},
	) {
		const labelTaskService = new LabelTaskService()
		const response = await labelTaskService.delete(new LabelTaskModel({
			taskId, labelId:
			label.id,
		}))
		const t = kanbanStore.getTaskById(taskId)
		if (t.task === null) {
			// Don't try further adding a label if the task is not in kanban
			// Usually this means the kanban board hasn't been accessed until now.
			// Vuex seems to have its difficulties with that, so we just log the error and fail silently.
			console.debug('Could not remove label from task in kanban, task not found', t)
			return response
		}

		// Remove the label from the project
		const labels = t.task.labels.filter(({ id }) => id !== label.id)

		kanbanStore.setTaskInBucketByIndex({
			...t,
			task: {
				...t.task,
				labels,
			},
		})

		return response
	}
	
	async function ensureLabelsExist(labels: string[]): Promise<LabelModel[]> {
		const all = [...new Set(labels)]
		const mustCreateLabel = all.map(async labelTitle => {
			let label = validateLabel(Object.values(labelStore.labels), labelTitle)
			if (typeof label === 'undefined') {
				// label not found, create it
				const labelModel = new LabelModel({
					title: labelTitle,
					hexColor: getRandomColorHex(),
				})
				label = await labelStore.createLabel(labelModel)
			}
			return label
		})
		return Promise.all(mustCreateLabel)
	}

	// Do everything that is involved in finding, creating and adding the label to the task
	async function addLabelsToTask(
		{ task, parsedLabels }:
		{ task: ITask, parsedLabels: string[] },
	) {
		if (parsedLabels.length <= 0) {
			return task
		}

		const labels = await ensureLabelsExist(parsedLabels)
		const labelAddsToWaitFor = labels.map(async l => addLabelToTask(task, l))

		// This waits until all labels are created and added to the task
		await Promise.all(labelAddsToWaitFor)
		return task
	}

	function findProjectId(
		{ project: projectName, projectId }:
		{ project: string, projectId: IProject['id'] }) {
		let foundProjectId = null

		// Uses the following ways to get the project id of the new task:
		//  1. If specified in quick add magic, look in store if it exists and use it if it does
		if (typeof projectName !== 'undefined' && projectName !== null) {
			let project = projectStore.findProjectByExactname(projectName)
			
			if (project === null) {
				project = projectStore.findProjectByIdentifier(projectName)
			}
			
			foundProjectId = project === null ? null : project.id
			if (foundProjectId !== null) {
				return foundProjectId
			}
		}
		
		//  2. Else check if a project was passed as parameter
		if (foundProjectId === null && projectId !== 0) {
			foundProjectId = projectId
		}
	
		//  3. Otherwise use the id from the route parameter
		const projectIdFromRoute = Number(router.currentRoute.value.params.projectId)
		if (typeof router.currentRoute.value.params.projectId !== 'undefined' && projectIdFromRoute > 0) {
			foundProjectId = projectIdFromRoute
		}
		
		//  4. If none of the above worked, reject the promise with an error.
		if (typeof foundProjectId === 'undefined' || projectId === null) {
			throw new Error('NO_PROJECT')
		}
	
		return foundProjectId
	}
	
	async function createNewTask({
		title,
		bucketId,
		projectId,
		position,
		index,
	} : 
		Partial<ITask>,
	) {
		const cancel = setModuleLoading(setIsLoading)
		const quickAddMagicMode = authStore.settings.frontendSettings.quickAddMagicMode
		const parsedTask = parseTaskText(title, quickAddMagicMode)

		if(parsedTask.text === '') {
			const taskService = new TaskService()
			try {
				return taskService.create(new TaskModel({
					title,
					projectId,
					bucketId,
					position,
					index,
				}))
			} finally {
				cancel()
			}
		}
	
		const foundProjectId = await findProjectId({
			project: parsedTask.project,
			projectId: projectId || 0,
		})
		
		if(foundProjectId === null || foundProjectId === 0) {
			cancel()
			throw new Error('NO_PROJECT')
		}

		const assignees = await findAssignees(parsedTask.assignees, foundProjectId)

		// Only clean up those assignees from the task title which actually exist
		let cleanedTitle = parsedTask.text
		if (assignees.length > 0) {
			const assigneePrefix = PREFIXES[quickAddMagicMode]?.assignee
			if (assigneePrefix) {
				cleanedTitle = cleanupItemText(cleanedTitle, assignees.map(a  => a.match), assigneePrefix)
			}
		}

		// I don't know why, but it all goes up in flames when I just pass in the date normally.
		const dueDate = parsedTask.date !== null ? new Date(parsedTask.date).toISOString() : null
	
		const task = new TaskModel({
			title: cleanedTitle,
			projectId: foundProjectId,
			dueDate,
			priority: parsedTask.priority,
			assignees,
			bucketId: bucketId || 0,
			position,
			index,
		})
		task.repeatAfter = parsedTask.repeats

		if (parsedTask.repeats?.type === REPEAT_TYPES.Months && parsedTask.repeats?.amount === 1) {
			task.repeatMode = TASK_REPEAT_MODES.REPEAT_MODE_MONTH
		}

		const taskService = new TaskService()
		try {
			const createdTask = await taskService.create(task)
			return await addLabelsToTask({
				task: createdTask,
				parsedLabels: parsedTask.labels,
			})
		} finally {
			cancel()
		}
	}
	
	async function setCoverImage(task: ITask, attachment: IAttachment | null) {
		return update({
			...task,
			coverImageAttachmentId: attachment ? attachment.id : 0,
		})
	}
	
	async function toggleFavorite(task: ITask) {
		const taskService = new TaskService()
		task.isFavorite = !task.isFavorite
		task = await taskService.update(task)
		
		// reloading the projects list so that the Favorites project shows up or is hidden when there are (or are not) favorite tasks
		await projectStore.loadAllProjects() 
		
		return task
	}

	return {
		tasks,
		isLoading,

		hasTasks,

		setTasks,
		loadTasks,
		update,
		delete: deleteTask, // since delete is a reserved word we have to alias here
		addTaskAttachment,
		addAssignee,
		removeAssignee,
		addLabel,
		removeLabel,
		addLabelsToTask,
		createNewTask,
		setCoverImage,
		findProjectId,
		ensureLabelsExist,
		toggleFavorite,
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useTaskStore, import.meta.hot))
}
