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
import {ref, useAttrs, watch, onBeforeUnmount, onMounted, nextTick} from 'vue'

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
	nextTick(() => {
		const dialog = dialogRef.value
		if (dialog) {
			delete dialog.dataset.closing
			dialog.showModal()
		}
		document.body.style.overflow = 'hidden'
	})
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
	{immediate: false},
)

onMounted(() => {
	if (props.enabled) {
		openDialog()
	}
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
$modal-mobile-margin: 1rem;

.modal-dialog {
	// Reset UA dialog styles
	padding: 0;
	border: none;
	background: transparent;
	color: #c0bdb8;
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
		background-color: rgba(3, 4, 8, .88);

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
	-webkit-overflow-scrolling: touch;
	padding-block-start: env(safe-area-inset-top);
	padding-block-end: max(env(safe-area-inset-bottom), 1rem);
	
	@media screen and (max-width: $tablet) {
		padding-inline-start: $modal-mobile-margin;
		padding-inline-end: $modal-mobile-margin;
	}
}

.modal-content,
.modal-card {
	background: var(--vk-bg-panel);
	border: .5px solid var(--vk-border-mid);
	border-radius: 12px;
	color: var(--vk-text-secondary);
	box-shadow: none;
	word-wrap: break-word;
	overflow-wrap: break-word;
}

.modal-header,
.modal-card-title {
	color: var(--vk-text-primary);
	word-break: break-word;
	hyphens: auto;
}

.modal-content .content,
.modal-card-body {
	color: var(--vk-text-secondary);
	word-wrap: break-word;
	overflow-wrap: break-word;
}

.modal-content .actions,
.modal-card-foot {
	border-block-start: .5px solid var(--vk-border-mid);
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
	
	@media screen and (max-width: $tablet) {
		flex-direction: column-reverse;
		gap: 0.75rem;
	}
}

.default .modal-content,
.hint-modal .modal-content {
	text-align: center;
	position: absolute;
	// fine to use top/left since we're only using this to position it centered
	inset-block-start: 50%;
	inset-inline-start: 50%;
	transform: translate(-50%, -50%);
	max-inline-size: 90vw;

	[dir="rtl"] & {
		transform: translate(50%, -50%);
	}

	@media screen and (max-width: $tablet) {
		margin: 0;
		position: static;
		transform: none;
		max-inline-size: 100%;
	}

	.modal-header {
		font-size: clamp(1.25rem, 5vw, 2rem);
		font-weight: 700;
		word-break: break-word;
	}

	.button {
		margin: 0 0.5rem;
		min-height: 44px;
		min-width: 44px;
		
		@media screen and (max-width: $tablet) {
			margin: 0.25rem 0;
		}
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
		max-inline-size: calc(100% - 2rem);
	}

	@media screen and (max-width: $tablet) {
		margin: 0.5rem;
		max-inline-size: calc(100% - 1rem);
		border-radius: 8px;
	}
}

.is-wide {
	max-inline-size: $desktop;
	inline-size: calc(100% - 2rem);
	
	@media screen and (max-width: $tablet) {
		max-inline-size: calc(100% - 1rem);
		inline-size: 100%;
	}
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
	color: var(--vk-text-secondary);
	font-size: clamp(1.5rem, 5vw, 2rem);
	transition: color .12s;
	padding: 8px;
	min-height: 44px;
	min-width: 44px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 8px;

	&:hover {
		color: var(--vk-accent-light);
		background-color: rgba(0, 0, 0, 0.1);
	}

	@media screen and (min-width: $desktop) and (width <= calc(#{$desktop	} + #{$close-button-min-space})) {
		inset-block-start: calc(5px + $modal-margin);
		inset-inline-end: 50%;
		// we align the close button to the modal until there is enough space outside for it
		transform: translateX(calc((#{$modal-width} / 2) - #{$close-button-padding}));
	}

	@media screen and (min-width: $tablet) and (max-width: #{$desktop + $close-button-min-space}) {
		inset-block-start: .75rem;
	}

	@media screen and (max-width: $tablet) {
		inset-block-start: max(0.5rem, env(safe-area-inset-top));
		inset-inline-end: max($close-button-padding, calc(env(safe-area-inset-right) + 0.5rem));
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
		padding-block-end: max(env(safe-area-inset-bottom), 1rem);
		padding-inline-start: $modal-mobile-margin;
		padding-inline-end: $modal-mobile-margin;
	}

	.modal-content {
		position: static;
		max-block-size: none;
		border-radius: 8px;
		margin-block: 0.5rem;
		padding: 1rem;
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
	color: var(--vk-text-secondary);
}

@media print, screen and (max-width: $tablet) {
  body:has(dialog[open].modal-dialog) #app {
	display: none;
  }
}
</style>
