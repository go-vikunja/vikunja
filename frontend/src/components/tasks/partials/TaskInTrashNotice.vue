<template>
	<InTrashNotice
		v-cy="'taskInTrashNotice'"
		:label="deletedLabel"
		:deleted-at="inTrash.deleted_at"
		:can-restore="inTrash.can_restore"
		:restoring="restoreTask.isPending.value"
		:trash-route="{name: 'trash.tasks'}"
		@restore="restoreTask.mutate(inTrash.id)"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import InTrashNotice from '@/components/misc/InTrashNotice.vue'
import {useRestoreTaskMutation, type TaskInTrash} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'

const props = defineProps<{
	inTrash: TaskInTrash,
}>()

const taskInTrashLabel = useDeletedLabel({
	unknown: 'trash.taskInTrash',
	by: 'trash.taskInTrashBy',
	byYou: 'trash.taskInTrashByYou',
})
const restoreTask = useRestoreTaskMutation()

const deletedLabel = computed(() => taskInTrashLabel(props.inTrash))
</script>
