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

<style lang="scss">
	.modal-mask {
		position: fixed;
		z-index: 9998;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-color: rgba(0, 0, 0, .8);
		transition: opacity .15s ease;
		color: #fff;

		.modal-container {
			transition: all .15s ease;
			position: relative;
			width: 100%;
			height: 100%;

			.modal-content {
				text-align: center;
				position: absolute;
				top: 50%;
				left: 50%;
				transform: translate(-50%, -50%);
				text-align: center;

				.header {
					font-size: 2rem;
					font-weight: 700;
				}

				.button {
					margin: 0 0.5rem;
				}
			}
		}
	}

	/* Transitions */

	.modal-enter {
		opacity: 0;
	}

	.modal-leave-active {
		opacity: 0;
	}

	.modal-enter .modal-container,
	.modal-leave-active .modal-container {
		-webkit-transform: scale(0.9);
		transform: scale(0.9);
	}
</style>
