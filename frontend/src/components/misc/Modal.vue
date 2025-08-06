<template>
	<Teleport to="body">
		<!-- FIXME: transition should not be included in the modal -->
		<CustomTransition
			:name="transitionName"
			appear
		>
			<section
				v-if="enabled"
				ref="modal"
				class="modal-mask"
				:class="[
					{ 'has-overflow': overflow },
					variant,
				]"
				v-bind="attrs"
			>
				<div
					v-shortcut="'Escape'"
					class="modal-container"
					@mousedown.self.prevent.stop="$emit('close')"
				>
					<BaseButton
						class="close"
						@click="$emit('close')"
					>
						<Icon icon="times" />
					</BaseButton>
					<div
						class="modal-content"
						:class="{
							'has-overflow': overflow,
							'is-wide': wide
						}"
					>
						<slot>
							<div class="modal-header">
								<slot name="header" />
							</div>
							<div class="content">
								<slot name="text" />
							</div>
							<div class="actions">
								<XButton
									variant="tertiary"
									class="has-text-danger"
									@click="$emit('close')"
								>
									{{ $t('misc.cancel') }}
								</XButton>
								<XButton
									v-cy="'modalPrimary'"
									variant="primary"
									:shadow="false"
									@click="$emit('submit')"
								>
									{{ $t('misc.doit') }}
								</XButton>
							</div>
						</slot>
					</div>
				</div>
			</section>
		</CustomTransition>
	</Teleport>
</template>

<script lang="ts" setup>
import CustomTransition from '@/components/misc/CustomTransition.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import {ref, useAttrs, watchEffect, onBeforeUnmount, watch} from 'vue'
import {useScrollLock} from '@vueuse/core'

const props = withDefaults(defineProps<{
	enabled?: boolean,
	overflow?: boolean,
	wide?: boolean,
	transitionName?: 'modal' | 'fade',
	variant?: 'default' | 'hint-modal' | 'scrolling',
}>(), {
	enabled: true,
	overflow: false,
	wide: false,
	transitionName: 'modal',
	variant: 'default',
})

const emit = defineEmits(['close', 'submit'])

defineOptions({
	inheritAttrs: false,
})

const attrs = useAttrs()

const modal = ref<HTMLElement | null>(null)
const scrollLock = useScrollLock(modal)

watchEffect(() => {
	scrollLock.value = props.enabled
})

function onKeydown(e: KeyboardEvent) {
	if (e.key === 'Escape') {
 		emit('close')
	}
}

watch(
	() => props.enabled,
	(value: boolean) => {
 		if (value) {
			window.addEventListener('keydown', onKeydown)
		} else {
			window.removeEventListener('keydown', onKeydown)
		}
	},
	{immediate: true},
)

onBeforeUnmount(() => {
	window.removeEventListener('keydown', onKeydown)
})
</script>

<style lang="scss" scoped>
$modal-margin: 4rem;
$modal-width: 1024px;

.modal-mask {
	position: fixed;
	z-index: 4000;
	inset-block-start: 0;
	inset-inline-start: 0;
	inline-size: 100%;
	block-size: 100%;
	background-color: rgba(0, 0, 0, .8);
	transition: opacity 150ms ease;
	color: #ffffff;
}

.modal-container {
	transition: all 150ms ease;
	position: relative;
	inline-size: 100%;
	block-size: 100%;
	max-block-size: 100vh;
	overflow: auto;
}

.default .modal-content,
.hint-modal .modal-content {
	text-align: center;
	position: absolute;
	// fine to use top/left since we're only using this to position it centered
	inset-block-start: 50%;
	inset-inline-start: 50%;
	transform: translate(-50%, -50%);

	@media screen and (max-width: $tablet) {
		margin: 0;
		position: static;
		transform: none;
	}

	.modal-header {
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
	inline-size: 100%;
	margin: $modal-margin auto;

	max-block-size: none; // reset bulma
	overflow: visible; // reset bulma
	
	@media not print {
		max-inline-size: $modal-width;
	}

	@media screen and (min-width: $tablet) {
		max-block-size: none; // reset bulma
		margin: $modal-margin auto; // reset bulma
		inline-size: 100%;
	}

	@media screen and (max-width: $desktop), print {
		margin: 0;
	}
}

.is-wide {
	max-inline-size: $desktop;
	inline-size: calc(100% - 2rem);
}

.hint-modal {
	z-index: 4600;

	:deep(.card-content) {
		text-align: start;

		.info {
			font-style: italic;
		}
	}
}

.close {
	$close-button-padding: 26px;
	position: fixed;
	inset-block-start: .5rem;
	inset-inline-end: $close-button-padding;
	color: var(--white);
	font-size: 2rem;

	@media screen and (min-width: $desktop) and (width <= calc(#{$desktop	} + #{$close-button-min-space})) {
		inset-block-start: calc(5px + $modal-margin);
		inset-inline-end: 50%;
		// we align the close button to the modal until there is enough space outside for it
		transform: translateX(calc((#{$modal-width} / 2) - #{$close-button-padding}));
	}

	@media screen and (min-width: $tablet) and (max-width: #{$desktop + $close-button-min-space}) {
		inset-block-start: .75rem;
	}
}

@media print, screen and (max-width: $tablet) {
  .modal-mask {
    overflow: visible !important;
  }

  .modal-container {
    block-size: auto;
	min-block-size: 100vh;
  }

  .modal-content {
    position: static;
	max-block-size: none;
  }

  .close {
    display: none;
  }

  :deep(.card) {
	border: none !important; 
	border-radius: 0 !important;
	min-block-size: 100vh;
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	margin-block-end: 0 !important;
  }
}

.modal-content:has(.modal-header) {
	display: flex;
	flex-direction: column;
	justify-content: center;
	padding: 0 1rem;
	min-block-size: 100vh
}

.modal-content :deep(.card .card-header-icon.close) {
	display: none;
	
	@media screen and (max-width: $tablet) {
		display: block;
	}
}
</style>

<style lang="scss">
// Close icon SVG uses currentColor, change the color to keep it visible
.dark .close {
	color: var(--grey-900);
}

@media print, screen and (max-width: $tablet) {
  body:has(.modal-mask) #app {
	display: none;
  }
}
</style>
