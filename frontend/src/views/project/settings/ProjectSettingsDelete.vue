<template>
	<Modal
		@close="goBack()"
		@submit="deleteProject()"
	>
		<template #header>
			<span>{{ $t('project.delete.header') }}</span>
		</template>

		<template #text>
			<template v-if="count">
				<p
					v-cy="'projectDeleteNotice'"
					class="tw:text-balance"
				>
					{{ trashNotice }}
				</p>
				<p
					v-if="count.time_entries"
					v-cy="'projectDeleteTimeEntries'"
					class="tw:text-balance has-text-weight-bold"
				>
					{{ $t('project.delete.timeEntries', {entries: count.time_entries, people: count.time_entry_authors}) }}
				</p>
			</template>
			<Loading
				v-else
				class="is-loading-small"
				variant="default"
			/>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'

import Loading from '@/components/misc/Loading.vue'
import {projectCountQuery, useDeleteProjectMutation} from '@/client/queries/projects'
import {useRestoreProjectMutation} from '@/client/queries/trash'
import {useGoBack} from '@/composables/useGoBack'
import {useProjects} from '@/composables/useProjects'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'

const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()
const deleteMutation = useDeleteProjectMutation()
const restoreMutation = useRestoreProjectMutation()
const route = useRoute()
const router = useRouter()
const goBack = useGoBack()

const projectId = computed(() => Number(route.params.projectId))
const project = computed(() => projectList.projects[projectId.value])
const count = useQuery(computed(() => projectCountQuery(projectId.value))).data

useTitle(() => t('project.delete.title', {project: project?.value?.title}))

const trashNotice = computed(() => {
	const subProjects = count.value?.sub_projects ?? 0
	const tasks = t('project.delete.tasks', count.value?.tasks ?? 0)
	return subProjects > 0
		? t('project.delete.toTrashWithSubProjects', {subProjects: t('project.delete.subProjects', subProjects), tasks})
		: t('project.delete.toTrash', {tasks})
})

async function deleteProject() {
	if (!project.value) {
		return
	}

	const id = project.value.id
	await deleteMutation.mutateAsync(id)
	// Works after unmount: the view navigates away before anyone clicks undo
	success({message: t('project.delete.success')}, [{
		title: t('task.undo'),
		callback: () => restoreMutation.mutate(id),
	}])
	router.push({name: 'home'})
}
</script>
