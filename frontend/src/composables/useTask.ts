import { ref, reactive, computed, type MaybeRefOrGetter, toValue, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { klona } from 'klona/lite'

import type { ITask } from '@/modelTypes/ITask'
import type { IProject } from '@/modelTypes/IProject'
import type { Priority } from '@/constants/priorities'
import type {Action as MessageAction} from '@/message'

import TaskModel from '@/models/task'

import { useTaskStore } from '@/stores/tasks'
import { useProjectStore } from '@/stores/projects'
import { useKanbanStore } from '@/stores/kanban'
import { useBaseStore } from '@/stores/base'

import { RIGHTS } from '@/constants/rights'
import { success } from '@/message'
import { playPopSound } from '@/helpers/playPop'
import { TASK_REPEAT_MODES } from '@/types/IRepeatMode'
import { shallowReactive } from 'vue'
import TaskService from '@/services/task'
import { useAttachmentStore } from '@/stores/attachments'
import { storeToRefs } from 'pinia'
import { getProjectTitle } from '@/helpers/getProjectTitle'
import { uploadFile } from '@/helpers/attachments'

export function useTask(taskId: MaybeRefOrGetter<ITask['id']>) {
	const {t} = useI18n()
	const router = useRouter()
	const taskStore = useTaskStore()

	const taskTitle = ref('')

	// We doubled the task color property here because verte does not have a real change property, leading
	// to the color property change being triggered when the # is removed from it, leading to an update,
	// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
	// updated, changed, updated and so on.
	// To prevent this, we put the task color property in a seperate value which is set to the task color
	// when it is saved and loaded.
	const taskColor = ref<ITask['hexColor']>('')

	const task = reactive({
		...new TaskModel(),
		title: taskTitle,
	}) as ITask

	const taskService = shallowReactive(new TaskService())

	/** Avoid flashing of empty elements if the task content is not yet loaded. */
	const isReady = ref(false)

	const isLoading = computed(() => taskService.loading)

	// load task
	watch(
		() => toValue(taskId),
		async (newTaskId) => {
			if (newTaskId === undefined) {
				return
			}
	
			try {
				const loaded = await taskService.get({ id: newTaskId })
				Object.assign(task, loaded)
				attachmentStore.set(task.attachments)
				taskColor.value = task.hexColor
				setActiveFields()
			} finally {
				isReady.value = true
			}
		},
		{immediate: true},
	)

	const canWrite = computed(() => (
		task.maxRight !== null &&
		task.maxRight > RIGHTS.READ
	))

	const projectStore = useProjectStore()

	const project = computed(() => projectStore.projects[task.projectId])

	const ancestorProjects = computed(() => projectStore.getAncestors(project.value))

	const ancestorProjectTitles = computed(() => ancestorProjects.value.map(project => getProjectTitle(project)))

	const attachmentStore = useAttachmentStore()
	const {hasAttachments} = storeToRefs(attachmentStore)

    async function saveTask(
        currentTask: ITask | null = null,
        undoCallback?: () => void,
    ) {
        if (currentTask === null) {
            currentTask = klona(task)
        }
    
        if (!canWrite.value) {
            return
        }
    
        currentTask.hexColor = taskColor.value
    
        // If no end date is being set, but a start date and due date,
        // use the due date as the end date
        if (
            currentTask.endDate === null &&
            currentTask.startDate !== null &&
            currentTask.dueDate !== null
        ) {
            currentTask.endDate = currentTask.dueDate
        }
    
        const updatedTask = await taskStore.update(currentTask) // TODO: markraw ?
        Object.assign(task, updatedTask)
        setActiveFields()
    
        let actions: MessageAction[] = []
        if (undoCallback) {
            actions = [{
                title: t('task.undo'),
                callback: undoCallback,
            }]
        }
        success({message: t('task.detail.updateSuccess')}, actions)
    }

	async function deleteTask() {
		await taskStore.delete(task)
		success({message: t('task.detail.deleteSuccess')})
		return router.push({name: 'project.index', params: {projectId: task.projectId}})
	}

	function uploadAttachment(file: File, onSuccess?: (url: string) => void) {
		return uploadFile(toValue(taskId), file, onSuccess)
	}

	async function toggleTaskDone() {
		const newTask = {
			...klona(task),
			done: !task.done,
		}
	
		if (newTask.done) {
			playPopSound()
		}
	
		await saveTask(
			newTask,
			toggleTaskDone,
		)
	}

	async function changeProject(project: IProject) {
		const kanbanStore = useKanbanStore()
		const baseStore = useBaseStore()
		kanbanStore.removeTaskInBucket(task)
		await saveTask({
			...task,
			projectId: project.id,
		})
		baseStore.setCurrentProject(project)
	}

	async function toggleFavorite() {
		const newTask = await taskStore.toggleFavorite(task)
		Object.assign(task, newTask)
	}

	async function setPriority(priority: Priority) {
		const newTask: ITask = {
			...task,
			priority,
		}
	
		return saveTask(newTask)
	}
	
	async function setPercentDone(percentDone: number) {
		const newTask: ITask = {
			...task,
			percentDone,
		}
	
		return saveTask(newTask)
	}
	
	async function removeRepeatAfter() {
		task.repeatAfter.amount = 0
		task.repeatMode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
		await saveTask()
	}

    return {
		project,
		ancestorProjects,
		ancestorProjectTitles,

		isReady,
		isLoading,

		task,
		taskTitle,
		taskColor,

		canWrite,
		hasAttachments,

		saveTask,
		deleteTask,
		uploadAttachment,
		
		toggleTaskDone,
		changeProject,
		toggleFavorite,
		setPriority,
		setPercentDone,
		removeRepeatAfter,
    }
}