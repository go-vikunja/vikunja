<template>
	<Modal
		@close="$router.back()"
		@submit="deleteProject()"
	>
		<template #header>
			<span>{{ $t('project.delete.header') }}</span>
		</template>

		<template #text>
			<p>
				{{ $t('project.delete.text1') }}
			</p>

			<p
				v-if="totalTasks !== null"
				class="has-text-weight-bold"
			>
				{{ deleteNotice }}
			</p>
			<Loading
				v-else
				class="is-loading-small"
				variant="default"
			/>

			<p>
				{{ $t('misc.cannotBeUndone') }}
			</p>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {useTitle} from '@/composables/useTitle'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {success} from '@/message'
import Loading from '@/components/misc/Loading.vue'
import {useProjectStore} from '@/stores/projects'
import TaskService from '@/services/task'
import {getRouteParamAsString} from '@/helpers/utils'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

const totalTasks = ref<number | null>(null)

const projectId = computed(() => {
	const id = getRouteParamAsString(route.params.projectId)
	return id ? parseInt(id, 10) : null
})

const project = computed(() => projectId.value ? projectStore.projects[projectId.value] : undefined)
const projectIdsToDelete = ref<number[]>([])

watchEffect(
	async () => {
		const currentProjectId = projectId.value
		if (!currentProjectId) {
			return
		}

		projectIdsToDelete.value = projectStore
			.getChildProjects(currentProjectId)
			.map(p => p.id)

		projectIdsToDelete.value.push(currentProjectId)

		const taskService = new TaskService()
		await taskService.getAll(undefined, {filter: `project in ${projectIdsToDelete.value.join(',')}`})
		totalTasks.value = taskService.totalPages * taskService.resultCount
	},
)

useTitle(() => t('project.delete.title', {project: project?.value?.title}))

const deleteNotice = computed(() => {
	if(totalTasks.value && totalTasks.value > 0) {
		if (projectIdsToDelete.value.length <= 1) {
			return t('project.delete.tasksToDelete', {count: totalTasks.value})
		} else if (projectIdsToDelete.value.length > 1) {
			return t('project.delete.tasksAndChildProjectsToDelete', {tasks: totalTasks.value, projects: projectIdsToDelete.value.length})
		}
	}

	return t('project.delete.noTasksToDelete')
})

async function deleteProject() {
	const projectToDelete = project.value
	if (!projectToDelete) {
		return
	}

	await projectStore.deleteProject(projectToDelete as IProject)
	success({message: t('project.delete.success')})
	router.push({name: 'home'})
}
</script>
