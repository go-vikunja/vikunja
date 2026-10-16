<template>
	<InTrashNotice
		v-cy="'projectInTrashNotice'"
		:label="deletedLabel"
		:deleted-at="inTrash.deleted_at"
		:can-restore="inTrash.can_restore"
		:restoring="restoreProject.isPending.value"
		:trash-route="{name: 'trash.projects'}"
		@restore="restoreProject.mutate(inTrash.id)"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import InTrashNotice from '@/components/misc/InTrashNotice.vue'
import {useRestoreProjectMutation, type ProjectInTrash} from '@/client/queries/trash'
import {useDeletedLabel} from '@/composables/useDeletedLabel'

const props = defineProps<{
	inTrash: ProjectInTrash,
}>()

const projectInTrashLabel = useDeletedLabel({
	unknown: 'trash.projectInTrash',
	by: 'trash.projectInTrashBy',
	byYou: 'trash.projectInTrashByYou',
})
const restoreProject = useRestoreProjectMutation()

const deletedLabel = computed(() => projectInTrashLabel(props.inTrash))
</script>
