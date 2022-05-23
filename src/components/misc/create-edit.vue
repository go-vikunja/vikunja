<template>
	<modal @close="$router.back()" :overflow="true" :wide="wide">
		<card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-left has-overflow"
			:has-close="true"
			@close="$router.back()"
			:loading="loading"
		>
			<div class="p-4">
				<slot></slot>
			</div>
			<footer class="modal-card-foot is-flex is-justify-content-flex-end">
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
					variant="primary"
					@click.prevent.stop="primary()"
					:icon="primaryIcon"
					:disabled="primaryDisabled"
				>
					{{ primaryLabel || $t('misc.create') }}
				</x-button>
			</footer>
		</card>
	</modal>
</template>

<script lang="ts">
</script>
<script setup lang="ts">
defineProps({
	title: {
		type: String,
		default: '',
	},
	primaryLabel: {
		type: String,
	},
	primaryIcon: {
		type: String,
		default: 'plus',
	},
	primaryDisabled: {
		type: Boolean,
		default: false,
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

const emit = defineEmits(['create', 'primary', 'tertiary'])

function primary() {
	emit('create')
	emit('primary')
}
</script>
