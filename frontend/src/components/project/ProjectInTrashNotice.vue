<template>
	<Message
		v-cy="'projectInTrashNotice'"
		variant="warning"
		class="project-in-trash"
	>
		<p>{{ deletedLabel }}</p>
		<p class="has-text-grey">
			{{ $t('trash.daysLeft', trashDaysLeft(inTrash.deleted_at)) }}
		</p>
		<div class="project-in-trash__actions">
			<XButton
				v-if="inTrash.can_restore"
				v-cy="'restoreTrashedProject'"
				icon="undo"
				:loading="restoreProject.isPending.value"
				@click="restoreProject.mutate(inTrash.id)"
			>
				{{ $t('trash.restore') }}
			</XButton>
			<XButton
				variant="secondary"
				:to="{name: 'trash.projects'}"
			>
				{{ $t('trash.open') }}
			</XButton>
		</div>
	</Message>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import Message from '@/components/misc/Message.vue'
import {trashDaysLeft, useRestoreProjectMutation, type ProjectInTrash} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'

const props = defineProps<{
	inTrash: ProjectInTrash,
}>()

const projectInTrashLabel = useDeletedLabel('trash.projectInTrash')
const restoreProject = useRestoreProjectMutation()

const deletedLabel = computed(() => projectInTrashLabel(props.inTrash))
</script>

<style lang="scss" scoped>
.project-in-trash__actions {
	display: flex;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
