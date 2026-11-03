<template>
	<Modal
		:enabled="counts !== null"
		@close="emit('cancel')"
		@submit="emit('confirm')"
	>
		<template #header>
			<span>{{ $t('project.share.accessLossConfirm.header') }}</span>
		</template>
		<template #text>
			<p v-if="counts">
				{{ $t('project.share.accessLossConfirm.text', {webhooks: counts.webhooks, linkShares: counts.linkShares}) }}
			</p>
			<ProjectAccessCleanupList
				v-model="cleanup"
				:intro="$t('project.share.accessLossConfirm.cleanupText')"
			/>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import Modal from '@/components/misc/Modal.vue'
import ProjectAccessCleanupList from '@/components/sharing/ProjectAccessCleanupList.vue'
import type {AccessLossCleanup} from '@/client/queries/accessLossCleanup'
import type {AccessLossCounts} from '@/composables/useAccessLossConfirmation'

defineProps<{
	counts: AccessLossCounts | null
}>()

const emit = defineEmits<{
	confirm: []
	cancel: []
}>()

const cleanup = defineModel<AccessLossCleanup>('cleanup', {required: true})

</script>
