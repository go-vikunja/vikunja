export const MIN_SCALE = 1
export const MAX_SCALE = 8

export interface ZoomTransform {
	scale: number
	translateX: number
	translateY: number
}

export interface ZoomMetrics {
	/** Layout size of the image, without the zoom transform applied. */
	imageWidth: number
	imageHeight: number
	/** Size of the box the image is centred in. */
	containerWidth: number
	containerHeight: number
	/** Viewport coordinates of the untransformed image centre. */
	centerX: number
	centerY: number
}

export interface ZoomAroundOptions {
	clientX: number
	clientY: number
	factor: number
}

function clamp(value: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, value))
}

export function clampScale(scale: number): number {
	return clamp(scale, MIN_SCALE, MAX_SCALE)
}

// Rough pixel equivalents of WheelEvent.deltaMode 0/1/2, so line- and page-based
// wheels (Firefox, some mice) land on the same sensitivity curve as pixel ones.
const WHEEL_PIXELS_PER_UNIT: Record<number, number> = {
	0: 1,
	1: 33,
	2: 400,
}
// Tuned so one classic wheel notch (~100px, or 3 lines) is roughly a 1.35x step.
const WHEEL_SENSITIVITY = 0.003
const WHEEL_FACTOR_LIMIT = 2

/**
 * Zoom factor for a single wheel event, proportional to its delta — a trackpad
 * emits dozens of small-delta events per flick and must not step like a notch.
 */
export function wheelZoomFactor(deltaY: number, deltaMode = 0): number {
	const pixels = deltaY * (WHEEL_PIXELS_PER_UNIT[deltaMode] ?? 1)
	const factor = Math.exp(-pixels * WHEEL_SENSITIVITY)

	return clamp(factor, 1 / WHEEL_FACTOR_LIMIT, WHEEL_FACTOR_LIMIT)
}

/**
 * Pins the image edges to the container edges: the overhang of the scaled image
 * is the whole pan budget, so an image that still fits cannot be panned at all.
 */
export function clampTranslate(transform: ZoomTransform, metrics: ZoomMetrics): ZoomTransform {
	const maxX = Math.max(0, (metrics.imageWidth * transform.scale - metrics.containerWidth) / 2)
	const maxY = Math.max(0, (metrics.imageHeight * transform.scale - metrics.containerHeight) / 2)

	return {
		scale: transform.scale,
		translateX: clamp(transform.translateX, -maxX, maxX),
		translateY: clamp(transform.translateY, -maxY, maxY),
	}
}

export function panBy(transform: ZoomTransform, metrics: ZoomMetrics, dx: number, dy: number): ZoomTransform {
	return clampTranslate({
		scale: transform.scale,
		translateX: transform.translateX + dx,
		translateY: transform.translateY + dy,
	}, metrics)
}

/** Zooms by `factor`, keeping the image point under the given viewport point put. */
export function zoomAround(
	transform: ZoomTransform,
	metrics: ZoomMetrics,
	{clientX, clientY, factor}: ZoomAroundOptions,
): ZoomTransform {
	const nextScale = clampScale(transform.scale * factor)
	if (nextScale === transform.scale) {
		return transform
	}

	if (nextScale === MIN_SCALE) {
		return {scale: nextScale, translateX: 0, translateY: 0}
	}

	// The rendered centre moves with the translation, so the anchor offset has to
	// be measured against it rather than against the layout centre.
	const offsetX = (clientX - (metrics.centerX + transform.translateX)) / transform.scale
	const offsetY = (clientY - (metrics.centerY + transform.translateY)) / transform.scale

	return clampTranslate({
		scale: nextScale,
		translateX: transform.translateX + offsetX * (transform.scale - nextScale),
		translateY: transform.translateY + offsetY * (transform.scale - nextScale),
	}, metrics)
}
