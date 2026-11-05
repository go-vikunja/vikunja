<template>
	<p class="tw:text-balance">{{ intro }}</p>
	<!-- Keep in sync with cleanupAfterProjectAccessLoss. -->
	<ul class="cleanup-list">
		<li>{{ $t(`project.share.accessCleanup.${perspective}.assignments`) }}</li>
		<li>{{ $t(`project.share.accessCleanup.${perspective}.subscriptions`) }}</li>
		<li v-if="perspective === 'self'">
			{{ $t('project.share.accessCleanup.self.webhooks') }}
		</li>
	</ul>
	<div class="cleanup-options">
		<FormCheckbox
			v-if="perspective === 'others'"
			:model-value="!cleanup.keep_webhooks"
			@update:modelValue="deleteWebhooks => cleanup = {...cleanup, keep_webhooks: !deleteWebhooks}"
		>
			{{ $t('project.share.accessCleanup.others.deleteWebhooks') }}
		</FormCheckbox>
		<FormCheckbox
			:model-value="cleanup.revoke_link_shares"
			@update:modelValue="revoke => cleanup = {...cleanup, revoke_link_shares: revoke}"
		>
			{{ $t(`project.share.accessCleanup.${perspective}.revokeLinkShares`) }}
		</FormCheckbox>
	</div>
</template>

<script setup lang="ts">
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import type {AccessLossCleanup} from '@/client/queries/accessLossCleanup'

const {perspective = 'others'} = defineProps<{
	intro: string
	perspective?: 'self' | 'others'
}>()

const cleanup = defineModel<AccessLossCleanup>({required: true})
</script>

<style lang="scss" scoped>
.cleanup-list {
	display: inline-block;
	text-align: start;
	margin-inline-start: 0;
	margin-block-start: 0;
	padding-inline-start: 1.25rem;
}

.cleanup-options {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: .5rem;
	margin-block-start: 1rem;
	text-align: start;
}
</style>
