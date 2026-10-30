import {describe, expect, it} from 'vitest'

import type {Label} from '@/client/generated'
import {
	groupTasksByVennRegion,
	labelsInUse,
	vennMaskOf,
	visibleMaskOf,
} from './vennRegions'

const b1: Label = {id: 1, title: 'Branch 1'}
const b2: Label = {id: 2, title: 'Branch 2'}
const b3: Label = {id: 3, title: 'Branch 3'}
const b4: Label = {id: 4, title: 'Branch 4'}

const p1 = {id: 11, labels: [b1, b2]}
const p2 = {id: 12, labels: [b2]}
const p3 = {id: 13, labels: [b3]}
const p4 = {id: 14, labels: [b1, b3]}
const p5 = {id: 15, labels: null}

describe('vennMaskOf', () => {
	it('sets one bit per selected label the task carries', () => {
		expect(vennMaskOf(p1, [1, 2, 3])).toBe(0b011)
		expect(vennMaskOf(p4, [1, 2, 3])).toBe(0b101)
		expect(vennMaskOf(p3, [1, 2])).toBe(0)
	})

	it('treats a task without labels as outside every set', () => {
		expect(vennMaskOf(p5, [1, 2, 3])).toBe(0)
	})

	it('ignores a label that appears twice on the same task', () => {
		expect(vennMaskOf({labels: [b2, b2]}, [1, 2])).toBe(0b10)
	})
})

describe('groupTasksByVennRegion', () => {
	it('puts every task in exactly one region', () => {
		const regions = groupTasksByVennRegion([p1, p2, p3, p4, p5], [1, 2, 3])

		expect(regions.get(0b011)).toEqual([p1])
		expect(regions.get(0b010)).toEqual([p2])
		expect(regions.get(0b100)).toEqual([p3])
		expect(regions.get(0b101)).toEqual([p4])
		expect(regions.get(0)).toEqual([p5])
		expect([...regions.values()].flat()).toHaveLength(5)
	})

	it('leaves out regions without tasks', () => {
		const regions = groupTasksByVennRegion([p1, p2, p3, p4, p5], [1, 2, 3])

		expect(regions.has(0b111)).toBe(false)
		expect(regions.size).toBe(5)
	})
})

describe('visibleMaskOf', () => {
	it('clears the bits of hidden labels', () => {
		expect(visibleMaskOf([1, 2, 3], [])).toBe(0b111)
		expect(visibleMaskOf([1, 2, 3], [2])).toBe(0b101)
		expect(visibleMaskOf([1, 2, 3], [9])).toBe(0b111)
	})
})

describe('labelsInUse', () => {
	it('orders labels by how many tasks use them, then by title', () => {
		expect(labelsInUse([p1, p2, p3, p4, p5]).map(l => l.id)).toEqual([1, 2, 3])
		expect(labelsInUse([{labels: [b3]}, {labels: [b4, b3]}, {labels: [b1]}]).map(l => l.id)).toEqual([3, 1, 4])
	})
})
