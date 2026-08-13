<template>
	<Modal
		:enabled="blobUrl !== null"
		:aria-label="$t('misc.imagePreview')"
		variant="fullscreen"
		@close="$emit('close')"
	>
		<div
			class="image-lightbox"
			@click.self="$emit('close')"
			@wheel.prevent="onWheel"
		>
			<div
				v-if="!loaded"
				class="image-lightbox__loader"
			>
				<Loading />
			</div>
			<img
				v-if="blobUrl !== null"
				ref="imageRef"
				:src="blobUrl"
				:alt="alt ?? ''"
				class="image-lightbox__image"
				:class="{
					'is-loaded': loaded,
					'is-zoomed': scale > 1,
					'is-panning': isPanning,
				}"
				:style="{transform: `translate(${translateX}px, ${translateY}px) scale(${scale})`}"
				draggable="false"
				@load="loaded = true"
				@dblclick="toggleZoom"
				@pointerdown="onPointerDown"
				@pointermove="onPointerMove"
				@pointerup="onPointerUp"
				@pointercancel="onPointerUp"
			>

			<div
				v-if="loaded"
				class="image-lightbox__toolbar"
			>
				<BaseButton
					:aria-label="$t('misc.zoomOut')"
					class="image-lightbox__button"
					@click="zoomByStep(1 / ZOOM_STEP)"
				>
					<Icon icon="minus" />
				</BaseButton>
				<span class="image-lightbox__level">{{ Math.round(scale * 100) }}%</span>
				<BaseButton
					:aria-label="$t('misc.zoomIn')"
					class="image-lightbox__button"
					@click="zoomByStep(ZOOM_STEP)"
				>
					<Icon icon="plus" />
				</BaseButton>
				<BaseButton
					:aria-label="$t('misc.resetZoom')"
					class="image-lightbox__button"
					@click="reset"
				>
					<Icon icon="undo" />
				</BaseButton>
			</div>
		</div>
	</Modal>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'

import Modal from '@/components/misc/Modal.vue'
import Loading from '@/components/misc/Loading.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps<{
	blobUrl: string | null,
	alt?: string,
}>()

defineEmits<{
	close: [],
}>()

const MIN_SCALE = 1
const MAX_SCALE = 8
const ZOOM_STEP = 1.4

const imageRef = ref<HTMLImageElement | null>(null)
const loaded = ref(false)

const scale = ref(1)
const translateX = ref(0)
const translateY = ref(0)

const isPanning = ref(false)
const pointers = new Map<number, {x: number, y: number}>()
let panStart = {x: 0, y: 0, translateX: 0, translateY: 0}
let pinchStartDistance = 0
let pinchStartScale = 1

// Start fresh whenever a new image is shown (or the lightbox reopens).
watch(() => props.blobUrl, () => {
	loaded.value = false
	reset()
})

function reset() {
	scale.value = 1
	translateX.value = 0
	translateY.value = 0
}

function clamp(value: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, value))
}

// Keep the scaled image from being dragged entirely out of view.
function clampPan() {
	const image = imageRef.value
	if (!image) {
		return
	}
	const maxX = (image.offsetWidth * scale.value) / 2
	const maxY = (image.offsetHeight * scale.value) / 2
	translateX.value = clamp(translateX.value, -maxX, maxX)
	translateY.value = clamp(translateY.value, -maxY, maxY)
}

// Zoom around a viewport point (cursor or pinch centre) so the pixel under it
// stays put.
function zoomAround(clientX: number, clientY: number, factor: number) {
	const image = imageRef.value
	if (!image) {
		return
	}
	const rect = image.getBoundingClientRect()
	const centerX = rect.left + rect.width / 2
	const centerY = rect.top + rect.height / 2
	const offsetX = (clientX - centerX) / scale.value
	const offsetY = (clientY - centerY) / scale.value

	const next = clamp(scale.value * factor, MIN_SCALE, MAX_SCALE)
	if (next === scale.value) {
		return
	}

	translateX.value += offsetX * (scale.value - next)
	translateY.value += offsetY * (scale.value - next)
	scale.value = next

	if (scale.value === MIN_SCALE) {
		translateX.value = 0
		translateY.value = 0
	} else {
		clampPan()
	}
}

function zoomByStep(factor: number) {
	zoomAround(window.innerWidth / 2, window.innerHeight / 2, factor)
}

function onWheel(event: WheelEvent) {
	zoomAround(event.clientX, event.clientY, event.deltaY < 0 ? ZOOM_STEP : 1 / ZOOM_STEP)
}

function toggleZoom(event: MouseEvent) {
	if (scale.value > MIN_SCALE) {
		reset()
	} else {
		zoomAround(event.clientX, event.clientY, 2.5)
	}
}

function pointerDistance(): number {
	const [a, b] = [...pointers.values()]
	return Math.hypot(a.x - b.x, a.y - b.y)
}

function onPointerDown(event: PointerEvent) {
	imageRef.value?.setPointerCapture(event.pointerId)
	pointers.set(event.pointerId, {x: event.clientX, y: event.clientY})

	if (pointers.size === 2) {
		pinchStartDistance = pointerDistance()
		pinchStartScale = scale.value
	} else if (pointers.size === 1 && scale.value > MIN_SCALE) {
		isPanning.value = true
		panStart = {
			x: event.clientX,
			y: event.clientY,
			translateX: translateX.value,
			translateY: translateY.value,
		}
	}
}

function onPointerMove(event: PointerEvent) {
	if (!pointers.has(event.pointerId)) {
		return
	}
	pointers.set(event.pointerId, {x: event.clientX, y: event.clientY})

	if (pointers.size === 2 && pinchStartDistance > 0) {
		const [a, b] = [...pointers.values()]
		const target = clamp(pinchStartScale * (pointerDistance() / pinchStartDistance), MIN_SCALE, MAX_SCALE)
		zoomAround((a.x + b.x) / 2, (a.y + b.y) / 2, target / scale.value)
	} else if (isPanning.value) {
		translateX.value = panStart.translateX + (event.clientX - panStart.x)
		translateY.value = panStart.translateY + (event.clientY - panStart.y)
		clampPan()
	}
}

function onPointerUp(event: PointerEvent) {
	pointers.delete(event.pointerId)
	if (pointers.size < 2) {
		pinchStartDistance = 0
	}
	if (pointers.size === 0) {
		isPanning.value = false
	}
}
</script>

<style scoped lang="scss">
.image-lightbox {
	position: relative;
	inline-size: 100%;
	block-size: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	touch-action: none;
}

.image-lightbox__loader {
	position: absolute;
	inset-block-start: 50%;
	inset-inline-start: 50%;
	transform: translate(-50%, -50%);
}

.image-lightbox__image {
	max-inline-size: 96vw;
	max-block-size: 90vh;
	object-fit: contain;
	user-select: none;
	-webkit-user-drag: none;
	cursor: zoom-in;
	opacity: 0;
	transition: opacity $transition;
	will-change: transform;

	&.is-loaded {
		opacity: 1;
	}

	&.is-zoomed {
		cursor: grab;
	}

	&.is-panning {
		cursor: grabbing;
		transition: none;
	}
}

.image-lightbox__toolbar {
	position: absolute;
	inset-block-end: 1.5rem;
	inset-inline-start: 50%;
	transform: translateX(-50%);
	display: flex;
	align-items: center;
	gap: .25rem;
	padding: .35rem;
	border-radius: 999px;
	background: hsla(var(--grey-900-hsl), .72);
	box-shadow: var(--shadow-md);
}

.image-lightbox__button {
	display: flex;
	align-items: center;
	justify-content: center;
	min-inline-size: 2.25rem;
	block-size: 2.25rem;
	padding: 0 .5rem;
	border-radius: 999px;
	color: var(--white);
	cursor: pointer;
	transition: background-color $transition;

	&:hover {
		background: rgba(255, 255, 255, .18);
	}
}

.image-lightbox__level {
	min-inline-size: 3rem;
	text-align: center;
	color: var(--white);
	font-size: .85rem;
	font-variant-numeric: tabular-nums;
	user-select: none;
}
</style>
