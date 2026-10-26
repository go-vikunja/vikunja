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
					{{ timeEntriesNotice }}
				</p>
			</template>
			<p
				v-else-if="countFailed"
				v-cy="'projectDeleteNotice'"
				class="tw:text-balance"
			>
				{{ $t('project.delete.toTrashFallback', {days: TRASH_RETENTION_DAYS}) }}
			</p>
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
import {projectCountQuery} from '@/client/queries/projects'
import {TRASH_RETENTION_DAYS, useDeleteProjectMutation, useRestoreProjectMutation} from '@/client/queries/trash'
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
const {data: count, isError: countFailed} = useQuery(computed(() => projectCountQuery(projectId.value)))

useTitle(() => t('project.delete.title', {project: project?.value?.title}))

const trashNotice = computed(() => {
	const subProjects = count.value?.sub_projects ?? 0
	const tasks = t('project.delete.tasks', count.value?.tasks ?? 0)
	return subProjects > 0
		? t('project.delete.toTrashWithSubProjects', {
			subProjects: t('project.delete.subProjects', subProjects),
			tasks,
			days: TRASH_RETENTION_DAYS,
		})
		: t('project.delete.toTrash', {
			tasks,
			days: TRASH_RETENTION_DAYS,
		})
})

const timeEntriesNotice = computed(() => t('project.delete.timeEntries', {
	entries: t('project.delete.timeEntryCount', count.value?.time_entries ?? 0),
	people: t('project.delete.people', count.value?.time_entry_authors ?? 0),
	days: TRASH_RETENTION_DAYS,
}))

async function deleteProject() {
	if (!project.value || deleteMutation.isPending.value) {
		return
	}

	const id = project.value.id
	try {
		await deleteMutation.mutateAsync(id)
	} catch {
		return
	}
	// Toast after navigating, so the undo can take the focus the closed view dropped.
	// Works after unmount: the view navigates away before anyone clicks undo
	await router.push({name: 'home'})
	success({message: t('project.delete.success')}, [{
		title: t('task.undo'),
		callback: () => restoreMutation.mutate(id),
	}])
}
</script>
