<template>
	<InTrashNotice
		v-cy="'projectInTrashNotice'"
		:label="deletedLabel"
		:deleted-at="inTrash.deleted_at"
		:can-restore="restoreTarget.allowed"
		:restore-label="restoreTarget.label"
		:restoring="restoreProject.isPending.value"
		:trash-route="{name: 'trash.projects'}"
		@restore="restoreProject.mutate(restoreTarget.id)"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

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
const {t} = useI18n({useScope: 'global'})

const deletedLabel = computed(() => projectInTrashLabel(props.inTrash))
// A trashed parent blocks restoring this project; restoring the root brings it back too
const restoreTarget = computed(() => props.inTrash.trashed_root_id
	? {
		id: props.inTrash.trashed_root_id,
		label: t('trash.restoreParentProject'),
		allowed: true,
	}
	: {
		id: props.inTrash.id,
		label: undefined,
		allowed: props.inTrash.can_restore,
	})
</script>
