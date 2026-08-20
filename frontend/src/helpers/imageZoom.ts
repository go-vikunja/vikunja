export const MIN_SCALE = 1
export const MAX_SCALE = 8

export interface ZoomTransform {
	scale: number
	translateX: number
	translateY: number
}

export interface ZoomMetrics {
	imageWidth: number
	imageHeight: number
	centerX: number
	centerY: number
}

function clamp(value: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, value))
}

export function clampScale(scale: number): number {
	return clamp(scale, MIN_SCALE, MAX_SCALE)
}

/** Keeps the scaled image from being dragged entirely out of view. */
export function clampTranslate(transform: ZoomTransform, metrics: ZoomMetrics): ZoomTransform {
	const maxX = (metrics.imageWidth * transform.scale) / 2
	const maxY = (metrics.imageHeight * transform.scale) / 2

	return {
		scale: transform.scale,
		translateX: clamp(transform.translateX, -maxX, maxX),
		translateY: clamp(transform.translateY, -maxY, maxY),
	}
}

/** Zooms by `factor`, keeping the image point under the given viewport point fixed. */
export function zoomAround(
	transform: ZoomTransform,
	metrics: ZoomMetrics,
	clientX: number,
	clientY: number,
	factor: number,
): ZoomTransform {
	const nextScale = clampScale(transform.scale * factor)
	if (nextScale === transform.scale) {
		return transform
	}

	if (nextScale === MIN_SCALE) {
		return {scale: nextScale, translateX: 0, translateY: 0}
	}

	// anchor against the rendered centre (it moves with the translation), not the layout centre
	const offsetX = (clientX - (metrics.centerX + transform.translateX)) / transform.scale
	const offsetY = (clientY - (metrics.centerY + transform.translateY)) / transform.scale

	return clampTranslate({
		scale: nextScale,
		translateX: transform.translateX + offsetX * (transform.scale - nextScale),
		translateY: transform.translateY + offsetY * (transform.scale - nextScale),
	}, metrics)
}
