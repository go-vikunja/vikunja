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
				v-if="tasksLoaded"
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
import {useDeleteProjectMutation} from '@/client/queries/projects'
import Loading from '@/components/misc/Loading.vue'
import {useProjects} from '@/composables/useProjects'
import {useTasks} from '@/composables/useTasks'

const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()
const deleteMutation = useDeleteProjectMutation()
const route = useRoute()
const router = useRouter()



const project = computed(() => projectList.projects[route.params.projectId])
const projectIdsToDelete = ref<number[]>([])

watchEffect(
	async () => {
		if (!route.params.projectId) {
			return
		}

		projectIdsToDelete.value = projectList
			.getChildProjects(parseInt(route.params.projectId))
			.map(p => p.id)

		projectIdsToDelete.value.push(parseInt(route.params.projectId))


	},
)

const taskQuery = useTasks(
	() => ({params: {filter: `project in ${projectIdsToDelete.value.join(',')}`, per_page: 1}}),
	{enabled: () => projectIdsToDelete.value.length > 0},
)
const totalTasks = taskQuery.total
const tasksLoaded = taskQuery.isSuccess

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
	if (!project.value) {
		return
	}

	await deleteMutation.mutateAsync(project.value.id)
	router.push({name: 'home'})
}
</script>
