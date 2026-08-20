import {describe, it, expect} from 'vitest'

import {
	MAX_SCALE,
	MIN_SCALE,
	clampScale,
	clampTranslate,
	panBy,
	wheelZoomFactor,
	zoomAround,
	type ZoomMetrics,
	type ZoomTransform,
} from './imageZoom'

// 800x600 image centred in a 1000x800 box that sits at the viewport origin.
const metrics: ZoomMetrics = {
	imageWidth: 800,
	imageHeight: 600,
	containerWidth: 1000,
	containerHeight: 800,
	centerX: 500,
	centerY: 400,
}

/** Where in image space the given viewport point lands under a transform. */
function imagePointUnder(transform: ZoomTransform, clientX: number, clientY: number) {
	return {
		x: (clientX - (metrics.centerX + transform.translateX)) / transform.scale,
		y: (clientY - (metrics.centerY + transform.translateY)) / transform.scale,
	}
}

describe('clampScale', () => {
	it('clamps to MIN_SCALE and MAX_SCALE exactly', () => {
		expect(clampScale(0.2)).toBe(MIN_SCALE)
		expect(clampScale(-5)).toBe(MIN_SCALE)
		expect(clampScale(100)).toBe(MAX_SCALE)
	})

	it('leaves values inside the range untouched', () => {
		expect(clampScale(1)).toBe(1)
		expect(clampScale(3.5)).toBe(3.5)
		expect(clampScale(8)).toBe(8)
	})
})

describe('zoomAround', () => {
	it('clamps at MAX_SCALE exactly', () => {
		const result = zoomAround({scale: 6, translateX: 0, translateY: 0}, metrics, {
			clientX: 500,
			clientY: 400,
			factor: 4,
		})

		expect(result.scale).toBe(MAX_SCALE)
	})

	it('returns to exactly 0,0 when the scale falls back to MIN_SCALE', () => {
		const result = zoomAround({scale: 2, translateX: 137, translateY: -89}, metrics, {
			clientX: 700,
			clientY: 500,
			factor: 0.1,
		})

		expect(result).toEqual({scale: MIN_SCALE, translateX: 0, translateY: 0})
	})

	it('does nothing when the scale is already at the bound', () => {
		const atMin: ZoomTransform = {scale: MIN_SCALE, translateX: 0, translateY: 0}
		expect(zoomAround(atMin, metrics, {clientX: 700, clientY: 500, factor: 0.5})).toEqual(atMin)

		const atMax: ZoomTransform = {scale: MAX_SCALE, translateX: 10, translateY: 10}
		expect(zoomAround(atMax, metrics, {clientX: 700, clientY: 500, factor: 2})).toEqual(atMax)
	})

	it('keeps the image point under the cursor fixed across a zoom step', () => {
		const clientX = 700
		const clientY = 500
		let transform: ZoomTransform = {scale: MIN_SCALE, translateX: 0, translateY: 0}
		const anchor = imagePointUnder(transform, clientX, clientY)

		expect(anchor).toEqual({x: 200, y: 100})

		for (const factor of [2, 1.5, 1.4, 1 / 1.4]) {
			transform = zoomAround(transform, metrics, {clientX, clientY, factor})
			const moved = imagePointUnder(transform, clientX, clientY)

			expect(moved.x).toBeCloseTo(anchor.x, 10)
			expect(moved.y).toBeCloseTo(anchor.y, 10)
		}

		expect(transform.scale).toBeCloseTo(2 * 1.5, 10)
	})
})

describe('panBy', () => {
	it('moves the image by the given step', () => {
		expect(panBy({scale: 2, translateX: 0, translateY: 0}, metrics, -48, 48))
			.toEqual({scale: 2, translateX: -48, translateY: 48})
	})

	it('stops at the container edge', () => {
		expect(panBy({scale: 2, translateX: 280, translateY: 0}, metrics, 48, 0).translateX).toBe(300)
	})

	it('cannot move an image that fits the container', () => {
		expect(panBy({scale: 1, translateX: 0, translateY: 0}, metrics, 48, 48))
			.toEqual({scale: 1, translateX: 0, translateY: 0})
	})
})

describe('wheelZoomFactor', () => {
	it('does not zoom without a delta', () => {
		expect(wheelZoomFactor(0, 0)).toBe(1)
	})

	it('turns one classic wheel notch into a 1.2x - 1.4x step', () => {
		const pixelNotch = wheelZoomFactor(-100, 0)
		const lineNotch = wheelZoomFactor(-3, 1)

		expect(pixelNotch).toBeGreaterThan(1.2)
		expect(pixelNotch).toBeLessThan(1.4)
		expect(lineNotch).toBeGreaterThan(1.2)
		expect(lineNotch).toBeLessThan(1.4)
	})

	it('stays proportional for the small deltas a trackpad emits', () => {
		// Ten 10px events must add up to the same zoom as a single 100px one.
		const flick = Array.from({length: 10}, () => wheelZoomFactor(-10, 0))
			.reduce((total, factor) => total * factor, 1)

		expect(flick).toBeCloseTo(wheelZoomFactor(-100, 0), 10)
	})

	it('is symmetric between scrolling up and down', () => {
		expect(wheelZoomFactor(-100, 0) * wheelZoomFactor(100, 0)).toBeCloseTo(1, 10)
	})

	it('caps a runaway delta at a 2x step', () => {
		expect(wheelZoomFactor(-10000, 0)).toBe(2)
		expect(wheelZoomFactor(10000, 0)).toBe(0.5)
	})
})

describe('clampTranslate', () => {
	it('allows no pan at all while the image fits the container', () => {
		const fitted = clampTranslate({scale: 1, translateX: 500, translateY: -500}, metrics)

		expect(fitted.scale).toBe(1)
		expect(fitted.translateX).toBeCloseTo(0, 10)
		expect(fitted.translateY).toBeCloseTo(0, 10)

		// 800 * 1.05 = 840, still inside the 1000 wide container.
		expect(clampTranslate({scale: 1.05, translateX: 400, translateY: 0}, metrics).translateX).toBe(0)
	})

	it('pins the image edges to the container edges once it overflows', () => {
		// 800 * 2 - 1000 = 600 of overhang, half of it on each side.
		expect(clampTranslate({scale: 2, translateX: 1000, translateY: -1000}, metrics))
			.toEqual({scale: 2, translateX: 300, translateY: -200})
	})

	it('leaves a translation inside the bounds untouched', () => {
		expect(clampTranslate({scale: 2, translateX: -120, translateY: 75}, metrics))
			.toEqual({scale: 2, translateX: -120, translateY: 75})
	})

	it('clamps each axis independently', () => {
		// The image overflows horizontally (1600 > 1000) but not vertically (600 < 800).
		const narrow: ZoomMetrics = {...metrics, imageHeight: 300}

		expect(clampTranslate({scale: 2, translateX: 250, translateY: 250}, narrow))
			.toEqual({scale: 2, translateX: 250, translateY: 0})
	})
})
