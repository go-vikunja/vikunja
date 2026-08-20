<template>
	<Modal
		:enabled="safeSrc !== null"
		:aria-label="$t('misc.imagePreview')"
		variant="fullscreen"
		@close="$emit('close')"
	>
		<div
			ref="containerRef"
			class="image-lightbox"
			@click.self="onBackdropClick"
			@wheel.prevent="onWheel"
		>
			<div
				v-if="!loaded && !failed"
				class="image-lightbox__loader"
			>
				<Loading />
			</div>
			<p
				v-if="failed"
				class="image-lightbox__error"
			>
				{{ $t('misc.imageLoadFailed') }}
			</p>
			<img
				v-if="safeSrc !== null && !failed"
				ref="imageRef"
				:src="safeSrc"
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
				@error="failed = true"
				@dblclick="toggleZoom"
				@pointerdown="onPointerDown"
				@pointermove="onPointerMove"
				@pointerup="onPointerUp"
				@pointercancel="onPointerUp"
				@lostpointercapture="onPointerUp"
			>

			<div
				v-if="loaded"
				class="image-lightbox__toolbar"
				@click.stop
				@pointerdown.stop
			>
				<BaseButton
					v-tooltip="$t('misc.zoomOut')"
					:aria-label="$t('misc.zoomOut')"
					class="image-lightbox__button"
					@click="zoomByStep(1 / ZOOM_STEP)"
				>
					<Icon icon="minus" />
				</BaseButton>
				<span
					class="image-lightbox__level"
					role="status"
					aria-live="polite"
				>{{ Math.round(scale * 100) }}%</span>
				<BaseButton
					v-tooltip="$t('misc.zoomIn')"
					:aria-label="$t('misc.zoomIn')"
					class="image-lightbox__button"
					@click="zoomByStep(ZOOM_STEP)"
				>
					<Icon icon="plus" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('misc.resetZoom')"
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
import {computed, ref, watch} from 'vue'

import Modal from '@/components/misc/Modal.vue'
import Loading from '@/components/misc/Loading.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {
	MIN_SCALE,
	clampScale,
	clampTranslate,
	zoomAround,
	type ZoomMetrics,
	type ZoomTransform,
} from '@/helpers/imageZoom'

const props = defineProps<{
	// An already-resolved object URL; null keeps the lightbox closed.
	blobUrl: string | null,
	alt?: string,
}>()

const emit = defineEmits<{
	close: [],
}>()

// A double click at the call site opens the lightbox on the first click; the
// second one lands on the fresh backdrop and would close it again.
function onBackdropClick(event: MouseEvent) {
	if (event.detail > 1) {
		return
	}
	emit('close')
}

// Never render a caller-supplied remote url as an image source.
const safeSrc = computed(() => props.blobUrl !== null && /^(blob:|data:image\/)/.test(props.blobUrl)
	? props.blobUrl
	: null)

const ZOOM_STEP = 1.4

const containerRef = ref<HTMLElement | null>(null)
const imageRef = ref<HTMLImageElement | null>(null)
const loaded = ref(false)
const failed = ref(false)

const scale = ref(1)
const translateX = ref(0)
const translateY = ref(0)

const isPanning = ref(false)
const pointers = new Map<number, {x: number, y: number}>()
let panStart = {x: 0, y: 0, translateX: 0, translateY: 0}
let pinchStartDistance = 0
let pinchStartScale = 1

// Closing mid-drag destroys the <img> before its pointerup, so the gesture state
// has to be dropped on both edges or the next open starts as a phantom pinch.
watch(safeSrc, src => {
	resetGestures()

	// Only reset the rest while opening: the Modal keeps rendering during its close
	// transition, so resetting on null flashes the loader over the fading scrim.
	if (src === null) {
		return
	}
	loaded.value = false
	failed.value = false
	reset()
})

function resetGestures() {
	pointers.clear()
	isPanning.value = false
	pinchStartDistance = 0
	pinchStartScale = 1
	panStart = {x: 0, y: 0, translateX: 0, translateY: 0}
}

function reset() {
	scale.value = 1
	translateX.value = 0
	translateY.value = 0
}

function currentTransform(): ZoomTransform {
	return {scale: scale.value, translateX: translateX.value, translateY: translateY.value}
}

function applyTransform(next: ZoomTransform) {
	scale.value = next.scale
	translateX.value = next.translateX
	translateY.value = next.translateY
}

function measure(): ZoomMetrics | null {
	const image = imageRef.value
	const container = containerRef.value
	if (!image || !container) {
		return null
	}
	const rect = container.getBoundingClientRect()
	return {
		imageWidth: image.offsetWidth,
		imageHeight: image.offsetHeight,
		containerWidth: container.clientWidth,
		containerHeight: container.clientHeight,
		centerX: rect.left + rect.width / 2,
		centerY: rect.top + rect.height / 2,
	}
}

function clampPan() {
	const metrics = measure()
	if (metrics === null) {
		return
	}
	applyTransform(clampTranslate(currentTransform(), metrics))
}

function zoomAt(clientX: number, clientY: number, factor: number) {
	const metrics = measure()
	if (metrics === null) {
		return
	}
	applyTransform(zoomAround(currentTransform(), metrics, {clientX, clientY, factor}))
}

function zoomByStep(factor: number) {
	const metrics = measure()
	if (metrics === null) {
		return
	}
	applyTransform(zoomAround(currentTransform(), metrics, {
		clientX: metrics.centerX,
		clientY: metrics.centerY,
		factor,
	}))
}

function onWheel(event: WheelEvent) {
	zoomAt(event.clientX, event.clientY, event.deltaY < 0 ? ZOOM_STEP : 1 / ZOOM_STEP)
}

function toggleZoom(event: MouseEvent) {
	if (scale.value > MIN_SCALE) {
		reset()
	} else {
		zoomAt(event.clientX, event.clientY, 2.5)
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
		const target = clampScale(pinchStartScale * (pointerDistance() / pinchStartDistance))
		zoomAt((a.x + b.x) / 2, (a.y + b.y) / 2, target / scale.value)
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
// Fills the parent Modal's fullscreen content box; the empty space around the
// image is the backdrop.
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

.image-lightbox__error {
	color: #ffffff;
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
	// Literal colors: the toolbar sits on the scrim, which never flips with the theme.
	background: rgba(0, 0, 0, .72);
	box-shadow: 0 10px 20px rgba(0, 0, 0, .3);
}

.image-lightbox__button {
	display: flex;
	align-items: center;
	justify-content: center;
	min-inline-size: 2.25rem;
	block-size: 2.25rem;
	padding: 0 .5rem;
	border-radius: 999px;
	color: #ffffff;
	cursor: pointer;
	transition: background-color $transition;

	&:hover {
		background: rgba(255, 255, 255, .18);
	}
}

.image-lightbox__level {
	min-inline-size: 3rem;
	text-align: center;
	color: #ffffff;
	font-size: .85rem;
	font-variant-numeric: tabular-nums;
	user-select: none;
}
</style>
