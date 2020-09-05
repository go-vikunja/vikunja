<template>
	<transition name="modal">
		<div class="modal-mask">
			<div class="modal-container">
				<div class="modal-content">
					<div class="header">
						<slot name="header"></slot>
					</div>
					<div class="content">
						<slot name="text"></slot>
					</div>
					<div class="actions">
						<button @click="$emit('close')" class="button is-danger is-inverted noshadow">Cancel</button>
						<button @click="$emit('submit')" class="button is-success noshadow">Do it!</button>
					</div>
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
}
</script>
