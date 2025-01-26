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
			@close="$router.back()"
		>
			<div class="p-4">
				<slot />
			</div>

			<template #footer>
				<slot name="footer">
					<x-button
						v-if="tertiary !== ''"
						:shadow="false"
						variant="tertiary"
						@click.prevent.stop="$emit('tertiary', $event)"
					>
						{{ tertiary }}
					</x-button>
					<x-button
						variant="secondary"
						@click.prevent.stop="$router.back()"
					>
						{{ $t('misc.cancel') }}
					</x-button>
					<x-button
						v-if="hasPrimaryAction"
						variant="primary"
						:icon="primaryIcon"
						:disabled="primaryDisabled || loading"
						class="ml-2"
						@click.prevent.stop="primary"
					>
						{{ primaryLabel || $t('misc.create') }}
					</x-button>
				</slot>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import type {FontAwesomeIconProps} from '@fortawesome/vue-fontawesome'

withDefaults(defineProps<{
	title: string,
	primaryLabel?: string,
	primaryIcon?: FontAwesomeIconProps['icon'],
	primaryDisabled?: boolean,
	hasPrimaryAction?: boolean,
	tertiary?: string,
	wide?: boolean,
	loading?: boolean
}>(), {
	primaryLabel: '',
	primaryIcon: 'plus',
	primaryDisabled: false,
	hasPrimaryAction: true,
	tertiary: '',
	wide: false,
	loading: false,
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
