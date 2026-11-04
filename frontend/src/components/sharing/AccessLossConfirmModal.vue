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
				{{ $t('project.share.accessLossConfirm.text', {created}) }}
			</p>
			<ProjectAccessCleanupList
				v-model="cleanup"
				:intro="$t('project.share.accessLossConfirm.cleanupText')"
			/>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import Modal from '@/components/misc/Modal.vue'
import ProjectAccessCleanupList from '@/components/sharing/ProjectAccessCleanupList.vue'
import type {AccessLossCleanup} from '@/client/queries/accessLossCleanup'
import type {AccessLossCounts} from '@/composables/useAccessLossConfirmation'

const props = defineProps<{
	counts: AccessLossCounts | null
}>()

const emit = defineEmits<{
	confirm: []
	cancel: []
}>()

const cleanup = defineModel<AccessLossCleanup>('cleanup', {required: true})

const {t} = useI18n({useScope: 'global'})

const created = computed(() => {
	if (!props.counts) return ''
	const {webhooks, linkShares} = props.counts
	const webhooksText = t('project.share.accessLossConfirm.webhooks', webhooks)
	const linkSharesText = t('project.share.accessLossConfirm.linkShares', linkShares)
	if (linkShares === 0) return webhooksText
	if (webhooks === 0) return linkSharesText
	return t('project.share.accessLossConfirm.webhooksAndLinkShares', {
		webhooks: webhooksText,
		linkShares: linkSharesText,
	})
})
</script>
