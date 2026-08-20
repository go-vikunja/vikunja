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
				:alt="alt || $t('misc.imagePreview')"
				tabindex="0"
				class="image-lightbox__image"
				:class="{
					'is-loaded': loaded,
					'is-zoomed': scale > 1,
					'is-panning': isPanning,
				}"
				:style="{transform: `translate(${translateX}px, ${translateY}px) scale(${scale})`}"
				draggable="false"
				@load="onLoad"
				@error="failed = true"
				@dblclick="toggleZoom"
				@keydown="onKeyDown"
				@pointerdown="onPointerDown"
				@pointermove="onPointerMove"
				@pointerup="onPointerUp"
				@pointercancel="onPointerUp"
				@lostpointercapture="onPointerUp"
			>

			<div
				v-if="loaded"
				class="image-lightbox__toolbar"
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
import {computed, onMounted, ref} from 'vue'
import {useEventListener} from '@vueuse/core'

import Modal from '@/components/misc/Modal.vue'
import Loading from '@/components/misc/Loading.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {
	MIN_SCALE,
	clampScale,
	clampTranslate,
	wheelZoomFactor,
	zoomAround,
	type ZoomMetrics,
	type ZoomTransform,
} from '@/helpers/imageZoom'

// Mounted only while open (callers gate it with v-if and key it by blobUrl, so a
// url swap remounts), which makes all per-open state the initial state here.
const props = defineProps<{
	// An already-resolved object URL.
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

// The blob-scheme check is the security boundary: a caller-supplied remote or
// data url must never end up as an image source.
const safeSrc = computed(() => props.blobUrl !== null && props.blobUrl.startsWith('blob:')
	? props.blobUrl
	: null)

// A rejected url renders nothing and can never emit close on its own, which would
// strand the caller in an open state.
onMounted(() => {
	if (props.blobUrl !== null && safeSrc.value === null) {
		emit('close')
	}
})

const ZOOM_STEP = 1.4

const containerRef = ref<HTMLElement | null>(null)
const imageRef = ref<HTMLImageElement | null>(null)
const loaded = ref(false)
const failed = ref(false)

const scale = ref(MIN_SCALE)
const translateX = ref(0)
const translateY = ref(0)

const isPanning = ref(false)
const pointers = new Map<number, {x: number, y: number}>()
let panStart = {x: 0, y: 0, translateX: 0, translateY: 0}
let pinchStartDistance = 0
let pinchStartScale = 1

function reset() {
	scale.value = MIN_SCALE
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

// A pending or failed image has a 0-sized box, which would turn any zoom or pan
// into a bogus transform.
const zoomable = computed(() => loaded.value && !failed.value)

// Measuring right after a transform write forces a reflow, so the untransformed
// box is cached and only re-read on load and on resize.
let metrics: ZoomMetrics | null = null

function refreshMetrics() {
	const image = imageRef.value
	const container = containerRef.value
	if (!zoomable.value || !image || !container || image.offsetWidth === 0) {
		metrics = null
		return
	}
	const rect = container.getBoundingClientRect()
	metrics = {
		imageWidth: image.offsetWidth,
		imageHeight: image.offsetHeight,
		containerWidth: container.clientWidth,
		containerHeight: container.clientHeight,
		centerX: rect.left + rect.width / 2,
		centerY: rect.top + rect.height / 2,
	}
}

function currentMetrics(): ZoomMetrics | null {
	if (metrics === null) {
		refreshMetrics()
	}
	return metrics
}

function onLoad() {
	loaded.value = true
	refreshMetrics()
}

// The 96vw/90vh box relayouts on resize while the transform persists, which can
// leave the image outside of it.
function onResize() {
	refreshMetrics()
	clampPan()
}

useEventListener(() => safeSrc.value === null ? null : window, 'resize', onResize)

function clampPan() {
	const measured = currentMetrics()
	if (measured === null) {
		return
	}
	applyTransform(clampTranslate(currentTransform(), measured))
}

function zoomAt(clientX: number, clientY: number, factor: number) {
	const measured = currentMetrics()
	if (measured === null) {
		return
	}
	applyTransform(zoomAround(currentTransform(), measured, clientX, clientY, factor))
}

function zoomByStep(factor: number) {
	const measured = currentMetrics()
	if (measured === null) {
		return
	}
	zoomAt(measured.centerX, measured.centerY, factor)
}

function onWheel(event: WheelEvent) {
	zoomAt(event.clientX, event.clientY, wheelZoomFactor(event.deltaY, event.deltaMode))
}

function toggleZoom(event: MouseEvent) {
	if (scale.value > MIN_SCALE) {
		reset()
	} else {
		zoomAt(event.clientX, event.clientY, 2.5)
	}
}

// Panning has to work without a pointer (WCAG 2.1.1); an arrow key moves the
// viewport, so the image travels the opposite way. A Map, not an object literal:
// that would also match prototype keys like 'constructor'.
const PAN_KEYS = new Map<string, {x: number, y: number}>([
	['ArrowLeft', {x: 1, y: 0}],
	['ArrowRight', {x: -1, y: 0}],
	['ArrowUp', {x: 0, y: 1}],
	['ArrowDown', {x: 0, y: -1}],
])
const PAN_STEP = 48

function onKeyDown(event: KeyboardEvent) {
	if (!zoomable.value) {
		return
	}

	// Ctrl/Cmd+'-' also arrives as key '-'; swallowing it would break browser zoom (WCAG 1.4.4).
	if (event.ctrlKey || event.metaKey || event.altKey) {
		return
	}

	switch (event.key) {
		case '+':
		case '=':
			event.preventDefault()
			zoomByStep(ZOOM_STEP)
			return
		case '-':
			event.preventDefault()
			zoomByStep(1 / ZOOM_STEP)
			return
		case '0':
			event.preventDefault()
			reset()
			return
	}

	const direction = PAN_KEYS.get(event.key)
	if (direction === undefined || scale.value <= MIN_SCALE) {
		return
	}
	const measured = currentMetrics()
	if (measured === null) {
		return
	}

	event.preventDefault()
	const transform = currentTransform()
	applyTransform(clampTranslate({
		scale: transform.scale,
		translateX: transform.translateX + direction.x * PAN_STEP,
		translateY: transform.translateY + direction.y * PAN_STEP,
	}, measured))
}

function pointerDistance(): number {
	const [a, b] = [...pointers.values()]
	return Math.hypot(a.x - b.x, a.y - b.y)
}

function onPointerDown(event: PointerEvent) {
	if (!zoomable.value) {
		return
	}
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
		return
	}
	if (pointers.size !== 1) {
		return
	}

	// Lifting one finger hands the gesture back to a pan: re-anchor on the
	// surviving pointer, otherwise the next move snaps to the pre-pinch offset.
	const [survivor] = [...pointers.values()]
	panStart = {
		x: survivor.x,
		y: survivor.y,
		translateX: translateX.value,
		translateY: translateY.value,
	}
	isPanning.value = scale.value > MIN_SCALE
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

	&:focus-visible {
		outline: 2px solid #ffffff;
		outline-offset: 2px;
	}

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
