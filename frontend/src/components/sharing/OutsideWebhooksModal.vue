<template>
	<Modal
		v-if="prompt"
		variant="hint-modal"
		:aria-label="$t('project.share.outsideWebhooks.title')"
		@close="prompt.answer(null)"
	>
		<Card
			class="has-no-shadow"
			:title="$t('project.share.outsideWebhooks.title')"
		>
			<p>{{ $t('project.share.outsideWebhooks.text', prompt.count) }}</p>
			<template #footer>
				<XButton
					variant="tertiary"
					@click="prompt.answer(null)"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="secondary"
					@click="prompt.answer(false)"
				>
					{{ $t('project.share.outsideWebhooks.keep') }}
				</XButton>
				<XButton
					danger
					@click="prompt.answer(true)"
				>
					{{ $t('project.share.outsideWebhooks.delete') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import type {OutsideWebhooksPrompt} from '@/composables/useOutsideWebhooksConfirmation'

defineProps<{
	prompt: OutsideWebhooksPrompt | null
}>()
</script>
