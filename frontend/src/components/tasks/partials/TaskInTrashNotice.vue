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

const emit = defineEmits<{
	restored: [],
}>()

const taskInTrashLabel = useDeletedLabel('trash.taskInTrash')
const projectList = useProjects()
const restoreTask = useRestoreTaskMutation()

const canRestore = computed(() => (projectList.projects[props.inTrash.project_id]?.max_permission ?? 0) > PERMISSIONS.READ)

const deletedLabel = computed(() => taskInTrashLabel(props.inTrash))

async function restore() {
	await restoreTask.mutateAsync(props.taskId)
	emit('restored')
}
</script>
