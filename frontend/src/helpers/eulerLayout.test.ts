import {describe, expect, it} from 'vitest'

import {layoutEuler, type Circle, type Point} from './eulerLayout'

const WIDTH = 600

function maskAt(circles: Circle[], p: Point) {
	return circles.reduce((mask, c, i) => Math.hypot(p.x - c.x, p.y - c.y) < c.r ? mask | (1 << i) : mask, 0)
}

function overlaps(a: Circle, b: Circle) {
	return Math.hypot(a.x - b.x, a.y - b.y) < a.r + b.r
}

function expectEveryRegionPlaced(layout: ReturnType<typeof layoutEuler>, masks: number[]) {
	expect(layout.unplaced).toEqual([])
	for (const mask of masks.filter(m => m !== 0)) {
		const anchor = layout.anchors.get(mask)
		expect(anchor, `region ${mask.toString(2)}`).toBeDefined()
		expect(maskAt(layout.circles, anchor!)).toBe(mask)
	}
}

function expectNoFalseOverlap(layout: ReturnType<typeof layoutEuler>, masks: number[]) {
	const {circles} = layout
	for (let i = 0; i < circles.length; i++) {
		for (let j = i + 1; j < circles.length; j++) {
			const shared = masks.some(mask => mask & (1 << i) && mask & (1 << j))
			if (!shared) {
				expect(overlaps(circles[i], circles[j]), `circles ${i} and ${j}`).toBe(false)
			}
		}
	}
}

// RTL Design, Verification, ISP, Career, Python, Docs as in the example project.
const SIX_SIZES = [5, 4, 4, 2, 3, 2]
const SIX = [1, 2, 4, 3, 5, 6, 7, 8, 9, 16, 32, 48, 0b010110]

describe('layoutEuler', () => {
	it('places every region of the example project inside exactly its circles', () => {
		const layout = layoutEuler(SIX_SIZES, SIX, WIDTH)

		expectEveryRegionPlaced(layout, SIX)
		expectNoFalseOverlap(layout, SIX)
	})

	it('keeps labels that share no task apart', () => {
		// Two labels share tasks; the other two stand alone.
		const masks = [1, 2, 3, 4, 8]
		const layout = layoutEuler([3, 2, 4, 1], masks, WIDTH)

		expectEveryRegionPlaced(layout, masks)
		expectNoFalseOverlap(layout, masks)
		expect(overlaps(layout.circles[0], layout.circles[1])).toBe(true)
	})

	it('gives labels with more tasks bigger circles, by area', () => {
		const layout = layoutEuler([9, 4, 1], [1, 2, 4], WIDTH)
		const [big, mid, small] = layout.circles

		expect(big.r).toBeGreaterThan(mid.r)
		expect(mid.r).toBeGreaterThan(small.r)
		expect(mid.r / big.r).toBeCloseTo(Math.sqrt(4 / 9), 5)
	})

	it('keeps a single-task label readable next to a big one', () => {
		const layout = layoutEuler([100, 1], [1, 2], WIDTH)

		expect(layout.circles[1].r / layout.circles[0].r).toBeGreaterThanOrEqual(0.32 - 1e-9)
	})

	it('keeps every circle and title on the canvas', () => {
		const layout = layoutEuler(SIX_SIZES, SIX, WIDTH)

		for (const c of layout.circles) {
			expect(c.x - c.r).toBeGreaterThanOrEqual(0)
			expect(c.x + c.r).toBeLessThanOrEqual(WIDTH)
			expect(c.y - c.r).toBeGreaterThanOrEqual(0)
			expect(c.y + c.r).toBeLessThanOrEqual(layout.height)
		}
		for (const title of layout.titles) {
			expect(title.y).toBeGreaterThan(0)
			expect(title.y).toBeLessThan(layout.height)
		}
	})

	it('puts every circle title outside its own circle', () => {
		const layout = layoutEuler(SIX_SIZES, SIX, WIDTH)

		layout.titles.forEach((title, i) => {
			const c = layout.circles[i]
			expect(Math.hypot(title.x - c.x, title.y - c.y)).toBeGreaterThan(c.r)
		})
	})

	it('keeps the outside spot below every circle', () => {
		const layout = layoutEuler(SIX_SIZES, SIX, WIDTH)

		for (const c of layout.circles) {
			expect(layout.outside.y - 30).toBeGreaterThan(c.y + c.r)
		}
	})

	it('places a chain of six labels', () => {
		const chain = [1, 2, 4, 8, 16, 32, 3, 6, 12, 24, 48]
		const layout = layoutEuler([3, 1, 2, 5, 2, 1], chain, WIDTH)

		expectEveryRegionPlaced(layout, chain)
		expectNoFalseOverlap(layout, chain)
	})

	it('handles a single label', () => {
		const layout = layoutEuler([3], [0, 1], WIDTH)

		expectEveryRegionPlaced(layout, [1])
		expect(layout.circles).toHaveLength(1)
	})

	it('is deterministic', () => {
		expect(layoutEuler(SIX_SIZES, SIX, WIDTH)).toEqual(layoutEuler(SIX_SIZES, SIX, WIDTH))
	})

	it('ignores the outside region', () => {
		const layout = layoutEuler([1, 1, 1, 1], [0, 1, 2, 4, 8], WIDTH)

		expect(layout.anchors.has(0)).toBe(false)
		expect(layout.unplaced).toEqual([])
	})
})
