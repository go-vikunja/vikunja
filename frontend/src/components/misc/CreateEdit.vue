<template>
	<Modal
		:overflow="true"
		:wide="wide"
		@close="$router.back()"
	>
		<Card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-left"
			:loading="loading"
			:show-close="true"
			@close="$router.back()"
		>
			<div class="p-4">
				<slot />
			</div>

			<template #footer>
				<slot name="footer">
					<XButton
						v-if="tertiary !== ''"
						:shadow="false"
						variant="tertiary"
						@click.prevent.stop="$emit('tertiary', $event)"
					>
						{{ tertiary }}
					</XButton>
					<XButton
						variant="secondary"
						@click.prevent.stop="$router.back()"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						v-if="hasPrimaryAction"
						variant="primary"
						:icon="primaryIcon"
						:disabled="primaryDisabled || loading"
						class="ml-2"
						@click.prevent.stop="primary"
					>
						{{ primaryLabel || $t('misc.create') }}
					</XButton>
				</slot>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import type {IconProp} from '@fortawesome/fontawesome-svg-core'
import {type PropType} from 'vue'

const _props = defineProps({
	title: {
		type: String,
		required: true,
	},
	primaryLabel: {
		type: String,
		default: '',
	},
	primaryIcon: {
		type: Object as PropType<IconProp>,
		default: () => ({ iconName: 'plus', prefix: 'fas' } as IconProp),
	},
	primaryDisabled: {
		type: Boolean,
		default: false,
	},
	hasPrimaryAction: {
		type: Boolean,
		default: true,
	},
	tertiary: {
		type: String,
		default: '',
	},
	wide: {
		type: Boolean,
		default: false,
	},
	loading: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits<{
	'create': [event: MouseEvent],
	'primary': [event: MouseEvent],
	'tertiary': [event: MouseEvent]
}>()

function primary(event: MouseEvent) {
	emit('create', event)
	emit('primary', event)
}
</script>
