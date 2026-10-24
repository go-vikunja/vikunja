<template>
	<InTrashNotice
		v-cy="'taskInTrashNotice'"
		:label="deletedLabel"
		:deleted-at="inTrash.deleted_at"
		:can-restore="canRestore"
		:restoring="restoreTask.isPending.value"
		:trash-route="{name: 'trash.tasks'}"
		restore-cy="restoreTrashedTask"
		@restore="restore()"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import InTrashNotice from '@/components/misc/InTrashNotice.vue'
import {useRestoreTaskMutation, type TaskInTrash} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {useProjects} from '@/composables/useProjects'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	taskId: number,
	inTrash: TaskInTrash,
}>()

const taskInTrashLabel = useDeletedLabel({
	unknown: 'trash.taskInTrash',
	by: 'trash.taskInTrashBy',
	byYou: 'trash.taskInTrashByYou',
})
const projectList = useProjects()
const restoreTask = useRestoreTaskMutation()

const canRestore = computed(() => {
	const project = projectList.projects[props.inTrash.project_id]
	return !project?.is_archived && (project?.max_permission ?? 0) > PERMISSIONS.READ
})

const deletedLabel = computed(() => taskInTrashLabel(props.inTrash))

function restore() {
	restoreTask.mutate(props.taskId)
}
</script>
