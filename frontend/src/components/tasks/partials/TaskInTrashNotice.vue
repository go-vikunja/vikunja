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
		<div class="task-in-trash__actions">
			<XButton
				v-if="canRestore"
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
import {useI18n} from 'vue-i18n'

import Message from '@/components/misc/Message.vue'
import {trashDaysLeft, useRestoreTaskMutation, type TaskInTrash} from '@/client/queries/trash'
import {useProjects} from '@/composables/useProjects'
import {PERMISSIONS} from '@/constants/permissions'
import {formatDateSince} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/helpers/user'

const props = defineProps<{
	taskId: number,
	inTrash: TaskInTrash,
}>()

const emit = defineEmits<{
	restored: [],
}>()

const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()
const restoreTask = useRestoreTaskMutation()

const canRestore = computed(() => (projectList.projects[props.inTrash.project_id]?.max_permission ?? 0) > PERMISSIONS.READ)

const deletedLabel = computed(() => {
	const since = formatDateSince(props.inTrash.deleted_at)
	return props.inTrash.deleted_by
		? t('trash.taskInTrashBy', {user: getDisplayName(props.inTrash.deleted_by), since})
		: t('trash.taskInTrash', {since})
})

async function restore() {
	await restoreTask.mutateAsync(props.taskId)
	emit('restored')
}
</script>

<style lang="scss" scoped>
.task-in-trash__actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
