<template>
	<modal
		:overflow="true"
		:wide="wide"
		@close="$router.back()"
	>
		<card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-left"
			:has-close="true"
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
						@click.prevent.stop="$emit('tertiary')"
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
						@click.prevent.stop="primary()"
					>
						{{ primaryLabel || $t('misc.create') }}
					</x-button>
				</slot>
			</template>
		</card>
	</modal>
</template>

<script setup lang="ts">
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

withDefaults(defineProps<{
	title: string,
	primaryLabel?: string,
	primaryIcon?: IconProp,
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

const emit = defineEmits(['create', 'primary', 'tertiary'])

function primary() {
	emit('create')
	emit('primary')
}
</script>
