<template>
	<transition name="modal">
		<div class="modal-mask">
			<div class="modal-container" @mousedown.self.prevent.stop="$emit('close')">
				<div class="modal-content" :class="{'has-overflow': overflow, 'is-wide': wide}">
					<slot>
						<div class="header">
							<slot name="header"></slot>
						</div>
						<div class="content">
							<slot name="text"></slot>
						</div>
						<div class="actions">
							<x-button
								@click="$emit('close')"
								type="tertary"
								class="has-text-danger"
							>
								{{ $t('misc.cancel') }}
							</x-button>
							<x-button
								@click="$emit('submit')"
								type="primary"
								:shadow="false"
							>
								{{ $t('misc.doit') }}
							</x-button>
						</div>
					</slot>
				</div>
			</div>
		</div>
	</transition>
</template>

<script>
export default {
	name: 'modal',
	mounted: function () {
		document.addEventListener('keydown', (e) => {
			// Close the model when escape is pressed
			if (e.keyCode === 27) {
				this.$emit('close')
			}
		})
	},
	props: {
		overflow: {
			type: Boolean,
			default: false,
		},
		wide: {
			type: Boolean,
			default: false,
		},
	},
}
</script>
