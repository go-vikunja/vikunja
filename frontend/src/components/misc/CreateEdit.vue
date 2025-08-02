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
			class="has-text-start"
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
						class="mis-2"
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
