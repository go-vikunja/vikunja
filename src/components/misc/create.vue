<template>
	<modal @close="$router.back()" :overflow="true">
		<card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-left has-overflow"
			:has-close="true"
			close-icon="times"
			@close="$router.back()"
		>
			<div class="p-4">
				<slot></slot>
			</div>
			<footer class="modal-card-foot is-flex is-justify-content-flex-end">
				<x-button
					:shadow="false"
					type="secondary"
					@click.prevent.stop="$router.back()"
				>
					Cancel
				</x-button>
				<x-button
					:shadow="false"
					type="primary"
					@click.prevent.stop="$emit('create')"
					icon="plus"
					:disabled="createDisabled"
				>
					{{ createLabel }}
				</x-button>
			</footer>
		</card>
	</modal>
</template>

<script>
export default {
	name: 'create',
	props: {
		title: {
			type: String,
			default: '',
		},
		createLabel: {
			type: String,
			default: 'Create',
		},
		createDisabled: {
			type: Boolean,
			default: false,
		},
	},
}
</script>
