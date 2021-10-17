<template>
	<transition name="modal">
		<section
			v-if="enabled"
			class="modal-mask"
			:class="[
				{ 'has-overflow': overflow },
				variant,
			]"
		>
			<div
				class="modal-container"
				:class="{'has-overflow': overflow}"
				@click.self.prevent.stop="$emit('close')"
				@shortkey="$emit('close')"
				v-shortkey="['esc']"
			>
				<div
					class="modal-content"
					:class="{
						'has-overflow': overflow,
						'is-wide': wide
					}"
				>
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
		</section>
	</transition>
</template>

<script>
export const TRANSITION_NAMES = {
	MODAL: 'modal',
	FADE: 'fade',
}

export const VARIANTS = {
	DEFAULT: 'default',
	HINT_MODAL: 'hint-modal',
	SCROLLING: 'scrolling',
}

function validValue(values) {
	return (value) => Object.values(values).includes(value)
}

export default {
	name: 'modal',
	mounted() {
		document.addEventListener('keydown', (e) => {
			// Close the model when escape is pressed
			if (e.keyCode === 27) {
				this.$emit('close')
			}
		})
	},
	props: {
		enabled: {
			type: Boolean,
			default: true,
		},
		overflow: {
			type: Boolean,
			default: false,
		},
		wide: {
			type: Boolean,
			default: false,
		},
		transitionName: {
			type: String,
			default: TRANSITION_NAMES.MODAL,
			validator: validValue(TRANSITION_NAMES),
		},
		variant: {
			type: String,
			default: VARIANTS.DEFAULT,
			validator: validValue(VARIANTS),
		},
	},
	emits: ['close', 'submit'],
}
</script>

<style lang="scss" scoped>
.modal-mask {
	position: fixed;
	z-index: 4000;
	top: 0;
	left: 0;
	width: 100%;
	height: 100%;
	background-color: rgba(0, 0, 0, .8);
	transition: opacity 150ms ease;
	color: #fff;
}

.modal-container {
	transition: all 150ms ease;
	position: relative;
	width: 100%;
	height: 100%;
	max-height: 100vh;
	overflow: auto;
}

.default .modal-content,
.hint-modal .modal-content {
	text-align: center;
	position: absolute;
	top: 50%;
	left: 50%;
	transform: translate(-50%, -50%);

	@media screen and (max-width: $tablet) {
		margin: 0;
		top: 25%;
		transform: translate(-50%, -25%);
	}

	.header {
		font-size: 2rem;
		font-weight: 700;
	}

	.button {
		margin: 0 0.5rem;
	}
}

// scrolling-content
// used e.g. for <TaskDetailViewModal>
.scrolling .modal-content {
	max-width: 1024px;
	width: 100%;
	margin: 4rem auto;

	max-height: none; // reset bulma
	overflow: visible; // reset bulma

	@media screen and (min-width: $tablet) {
		max-height: none; // reset bulma
		margin: 4rem auto; // reset bulma
		width: 100%;
	}

	@media screen and (max-width: $desktop) {
		margin: 0;
	}
}

.is-wide {
	max-width: $desktop;
	width: calc(100% - 2rem);
}

.hint-modal {
	z-index: 4600;

	::v-deep.card-content {
		text-align: left;

		.info {
			font-style: italic;
		}

		p {
			display: flex;
			justify-content: space-between;
			align-items: center;

		}

		.message-body {
			padding: .5rem .75rem;
		}
	}
}



/* Transitions */

.modal-enter,
.modal-leave-active {
  opacity: 0;
}

.modal-enter .modal-container,
.modal-leave-active .modal-container {
  transform: scale(0.9);
}
</style>