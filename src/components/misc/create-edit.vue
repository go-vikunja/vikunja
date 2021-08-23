<template>
	<modal @close="$router.back()" :overflow="true" :wide="wide">
		<card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-left has-overflow"
			:has-close="true"
			close-icon="times"
			@close="$router.back()"
			:loading="loading"
		>
			<div class="p-4">
				<slot></slot>
			</div>
			<footer class="modal-card-foot is-flex is-justify-content-flex-end">
				<x-button
					:shadow="false"
					type="tertary"
					@click.prevent.stop="$emit('tertary')"
					v-if="tertary !== ''"
				>
					{{ tertary }}
				</x-button>
				<x-button
					type="secondary"
					@click.prevent.stop="$router.back()"
				>
					{{ $t('misc.cancel') }}
				</x-button>
				<x-button
					type="primary"
					@click.prevent.stop="primary"
					:icon="primaryIcon"
					:disabled="primaryDisabled"
					v-if="primaryLabel !== ''"
				>
					{{ primaryLabel }}
				</x-button>
			</footer>
		</card>
	</modal>
</template>

<script>
export default {
	name: 'create-edit',
	props: {
		title: {
			type: String,
			default: '',
		},
		primaryLabel: {
			type: String,
			default() {
				return this.$t('misc.create')
			},
		},
		primaryIcon: {
			type: String,
			default: 'plus',
		},
		primaryDisabled: {
			type: Boolean,
			default: false,
		},
		tertary: {
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
	},
	emits: ['create', 'primary', 'tertary'],
	methods: {
		primary() {
			this.$emit('create')
			this.$emit('primary')
		},
	},
}
</script>
