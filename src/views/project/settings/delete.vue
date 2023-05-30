<template>
	<modal
		@close="$router.back()"
		@submit="deleteProject()"
	>
		<template #header><span>{{ $t('project.delete.header') }}</span></template>

		<template #text>
			<p>
				{{ $t('project.delete.text1') }}
			</p>

			<p>
				<strong v-if="totalTasks !== null" class="has-text-white">
					{{
						totalTasks > 0 ? $t('project.delete.tasksToDelete', {count: totalTasks}) : $t('project.delete.noTasksToDelete')
					}}
				</strong>
				<Loading v-else class="is-loading-small" variant="default"/>
			</p>

			<p>
				{{ $t('misc.cannotBeUndone') }}
			</p>
		</template>
	</modal>
</template>

<script setup lang="ts">
import {computed, ref, watchEffect} from 'vue'
import {useTitle} from '@/composables/useTitle'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {success} from '@/message'
import TaskCollectionService from '@/services/taskCollection'
import Loading from '@/components/misc/loading.vue'
import {useProjectStore} from '@/stores/projects'

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

const totalTasks = ref<number | null>(null)

const project = computed(() => projectStore.projects[route.params.projectId])

watchEffect(
	() => {
		if (!route.params.projectId) {
			return
		}

		const taskCollectionService = new TaskCollectionService()
		taskCollectionService.getAll({projectId: route.params.projectId}).then(() => {
			totalTasks.value = taskCollectionService.totalPages * taskCollectionService.resultCount
		})
	},
)

useTitle(() => t('project.delete.title', {project: project?.value?.title}))

async function deleteProject() {
	if (!project.value) {
		return
	}

	await projectStore.deleteProject(project.value)
	success({message: t('project.delete.success')})
	router.push({name: 'home'})
}
</script>
