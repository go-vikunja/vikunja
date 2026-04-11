<template>
	<Teleport to="body">
		<dialog
			v-if="showDialog"
			ref="dialogRef"
			class="modal-dialog"
			:class="[
				{ 'has-overflow': overflow },
				variant,
			]"
			v-bind="attrs"
			@cancel.prevent="$emit('close')"
		>
			<div
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
		</dialog>
	</Teleport>
</template>

<script lang="ts" setup>
import BaseButton from '@/components/base/BaseButton.vue'
import {ref, useAttrs, watch, onBeforeUnmount} from 'vue'

const props = withDefaults(defineProps<{
	enabled?: boolean,
	overflow?: boolean,
	wide?: boolean,
	variant?: 'default' | 'hint-modal' | 'scrolling',
}>(), {
	enabled: true,
	overflow: false,
	wide: false,
	variant: 'default',
})

defineEmits(['close', 'submit'])

defineOptions({
	inheritAttrs: false,
})

const TRANSITION_DURATION = 150

const attrs = useAttrs()
const dialogRef = ref<HTMLDialogElement | null>(null)
const previouslyFocused = ref<Element | null>(null)
const showDialog = ref(false)
let closeTimer: ReturnType<typeof setTimeout> | null = null

function openDialog() {
	if (closeTimer) {
		clearTimeout(closeTimer)
		closeTimer = null
	}
	previouslyFocused.value = document.activeElement
	showDialog.value = true
	document.body.style.overflow = 'hidden'
	// If we're re-opening while the previous close transition is still in
	// flight the <dialog> is still mounted and [open], so the dialogRef
	// watcher below won't re-fire. Clear the data-closing flag here so the
	// dialog doesn't stay stuck at opacity 0.
	if (dialogRef.value) {
		delete dialogRef.value.dataset.closing
	}
	// The initial `showModal()` call happens in the `watch(dialogRef, …)`
	// below, which fires the moment Vue mounts the <dialog>. We cannot call
	// it synchronously here because the element is not in the DOM yet
	// (v-if="showDialog" only just became true), and we cannot rely on a
	// single nextTick because the mount can be deferred past it (#2590).
}

function closeDialog() {
	const dialog = dialogRef.value
	if (!dialog) return

	// Trigger the fade-out while the dialog is still [open] so the opacity
	// transition plays in browsers that don't support allow-discrete (Firefox).
	dialog.dataset.closing = ''
	document.body.style.overflow = ''

	closeTimer = setTimeout(() => {
		delete dialog.dataset.closing
		dialog.close()
		showDialog.value = false
		closeTimer = null
		if (previouslyFocused.value instanceof HTMLElement) {
			previouslyFocused.value.focus()
		}
		previouslyFocused.value = null
	}, TRANSITION_DURATION)
}

watch(
	() => props.enabled,
	(isEnabled) => {
		if (isEnabled) {
			openDialog()
		} else {
			closeDialog()
		}
	},
	{immediate: true},
)

// Actually call showModal() the moment the <dialog> element is mounted.
// `dialogRef` is populated by Vue during the render flush after
// `showDialog.value = true`, so this fires deterministically, no matter
// how many flushes the renderer needs (see #2590). We re-check
// `props.enabled` here because the prop can flip back to `false` between
// `openDialog()` and the mount flush, in which case we must not open.
watch(dialogRef, (dialog) => {
	if (!dialog) return
	if (!props.enabled) return
	delete dialog.dataset.closing
	dialog.showModal()
})

onBeforeUnmount(() => {
	if (closeTimer) {
		clearTimeout(closeTimer)
		closeTimer = null
	}
	document.body.style.overflow = ''
	if (previouslyFocused.value instanceof HTMLElement) {
		previouslyFocused.value.focus()
	}
})
</script>

<style lang="scss" scoped>
$modal-margin: 4rem;
$modal-width: 1024px;

.modal-dialog {
	// Reset UA dialog styles
	padding: 0;
	border: none;
	background: transparent;
	color: #ffffff;
	// Fill viewport
	position: fixed;
	inset: 0;
	inline-size: 100%;
	block-size: 100%;
	max-inline-size: 100%;
	max-block-size: 100%;

	// Transitions
	opacity: 0;
	transition: opacity 150ms ease,
				display 150ms ease allow-discrete;

	&[open]:not([data-closing]) {
		opacity: 1;

		@starting-style {
			opacity: 0;
		}
	}

	&::backdrop {
		background-color: rgba(0, 0, 0, 0);
		transition: background-color 150ms ease,
					display 150ms ease allow-discrete;
	}

	&[open]:not([data-closing])::backdrop {
		background-color: rgba(0, 0, 0, .8);

		@starting-style {
			background-color: rgba(0, 0, 0, 0);
		}
	}
}

.modal-container {
	position: relative;
	inline-size: 100%;
	block-size: 100%;
	max-block-size: 100dvh;
	overflow: auto;
	padding-block-start: env(safe-area-inset-top);
	padding-block-end: env(safe-area-inset-bottom);

}

.default .modal-content,
.hint-modal .modal-content {
	text-align: center;
	position: absolute;
	// fine to use top/left since we're only using this to position it centered
	inset-block-start: 50%;
	inset-inline-start: 50%;
	transform: translate(-50%, -50%);

	[dir="rtl"] & {
		transform: translate(50%, -50%);
	}

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
	.modal-dialog {
		overflow: visible !important;
	}

	.modal-container {
		block-size: auto;
		min-block-size: 100dvh;
		padding-block-start: env(safe-area-inset-top);
		padding-block-end: env(safe-area-inset-bottom);
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
		min-block-size: calc(100dvh - env(safe-area-inset-top) - env(safe-area-inset-bottom));
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
	min-block-size: calc(100dvh - env(safe-area-inset-top) - env(safe-area-inset-bottom));
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
.dark .modal-dialog .close {
	color: var(--grey-900);
}

@media print, screen and (max-width: $tablet) {
  body:has(dialog[open].modal-dialog) #app {
	display: none;
  }
}
</style>
