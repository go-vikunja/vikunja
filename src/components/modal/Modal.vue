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
						<button class="button is-danger is-inverted noshadow" @click="$emit('close')">Cancel</button>
						<button class="button is-success noshadow" @click="$emit('submit')">Do it!</button>
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
		}
	}
</script>
