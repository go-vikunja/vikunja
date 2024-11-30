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
import {computed, shallowReactive, watchEffect} from 'vue'
import {useTitle} from '@/composables/useTitle'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {success} from '@/message'
import Loading from '@/components/misc/Loading.vue'
import {useProjectStore} from '@/stores/projects'
import TaskService from '@/services/task'

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

const projectId = computed(() => Number(route.params.projectId))

const project = computed(() => projectStore.projects[projectId.value])

const projectIdsToDelete = computed(() => {
	if (!projectId.value) {
		return []
	}

	return [
		...projectStore
			.getChildProjects(projectId.value)
			.map(p => p.id),
		projectId.value,
	]
})

const taskService = shallowReactive(new TaskService())
watchEffect(
	async () => {
		if (!projectIdsToDelete.value.length) {
			return
		}


		await taskService.getAll({}, {filter: `project in ${projectIdsToDelete.value.join(',')}`})
	},
)
const totalTasks = computed(() => taskService.totalPages * taskService.resultCount)

useTitle(() => t('project.delete.title', {project: project?.value?.title}))

const deleteNotice = computed(() => {
	if (totalTasks.value && totalTasks.value > 0) {
		if (projectIdsToDelete.value.length <= 1) {
			return t('project.delete.tasksToDelete', {count: totalTasks.value})
		}
		
		if (projectIdsToDelete.value.length > 1) {
			return t('project.delete.tasksAndChildProjectsToDelete', {tasks: totalTasks.value, projects: projectIdsToDelete.value.length})
		}
	}

	return t('project.delete.noTasksToDelete')
})

async function deleteProject() {
	if (!project.value) {
		return
	}

	await projectStore.deleteProject(project.value)
	success({message: t('project.delete.success')})
	router.push({name: 'home'})
}
</script>
