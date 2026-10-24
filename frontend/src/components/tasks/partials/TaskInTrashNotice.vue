<template>
	<Message
		v-cy="'taskInTrashNotice'"
		variant="warning"
		class="task-in-trash"
	>
		<p>{{ deletedLabel }}</p>
		<p class="has-text-grey">
			{{ $t('trash.daysLeft', trashDaysLeft(inTrash.deleted_at)) }}
		</p>
		<!-- The trash only lists what the user can restore -->
		<div
			v-if="canRestore"
			class="task-in-trash__actions"
		>
			<XButton
				v-cy="'restoreTrashedTask'"
				icon="undo"
				:loading="restoreTask.isPending.value"
				@click="restore()"
			>
				{{ $t('trash.restore') }}
			</XButton>
			<XButton
				variant="secondary"
				:to="{name: 'trash.tasks'}"
			>
				{{ $t('trash.open') }}
			</XButton>
		</div>
	</Message>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import Message from '@/components/misc/Message.vue'
import {trashDaysLeft, useRestoreTaskMutation, type TaskInTrash} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'
import {useProjects} from '@/composables/useProjects'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	taskId: number,
	inTrash: TaskInTrash,
}>()

const taskInTrashLabel = useDeletedLabel('trash.taskInTrash')
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

<style lang="scss" scoped>
.task-in-trash__actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
