import {test, expect} from 'vitest'

import {calculateNearestHours} from './calculateNearestHours'

test('5:00', () => {
	const date = new Date()
	date.setHours(5)
	expect(calculateNearestHours(date)).toBe(9)
})

test('7:00', () => {
	const date = new Date()
	date.setHours(7)
	expect(calculateNearestHours(date)).toBe(9)
})

test('7:41', () => {
	const date = new Date()
	date.setHours(7)
	date.setMinutes(41)
	expect(calculateNearestHours(date)).toBe(9)
})

test('9:00', () => {
	const date = new Date()
	date.setHours(9)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(9)
})

test('10:00', () => {
	const date = new Date()
	date.setHours(10)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(12)
})

test('12:00', () => {
	const date = new Date()
	date.setHours(12)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(12)
})

test('13:00', () => {
	const date = new Date()
	date.setHours(13)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(15)
})

test('15:00', () => {
	const date = new Date()
	date.setHours(15)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(15)
})

test('16:00', () => {
	const date = new Date()
	date.setHours(16)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(18)
})

test('18:00', () => {
	const date = new Date()
	date.setHours(18)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(18)
})

test('19:00', () => {
	const date = new Date()
	date.setHours(19)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(21)
})

test('22:00', () => {
	const date = new Date()
	date.setHours(22)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(9)
})

test('22:40', () => {
	const date = new Date()
	date.setHours(22)
	date.setMinutes(0)
	expect(calculateNearestHours(date)).toBe(9)
})

// Test cases for the bug: when current time is past a breakpoint hour
test('12:30 should return next breakpoint (15), not current (12)', () => {
	const date = new Date()
	date.setHours(12)
	date.setMinutes(30)
	expect(calculateNearestHours(date)).toBe(15)
})

test('15:54 should return next breakpoint (18), not current (15)', () => {
	const date = new Date()
	date.setHours(15)
	date.setMinutes(54)
	expect(calculateNearestHours(date)).toBe(18)
})

test('18:45 should return next breakpoint (21), not current (18)', () => {
	const date = new Date()
	date.setHours(18)
	date.setMinutes(45)
	expect(calculateNearestHours(date)).toBe(21)
})

test('21:30 should return next day breakpoint (9), not current (21)', () => {
	const date = new Date()
	date.setHours(21)
	date.setMinutes(30)
	expect(calculateNearestHours(date)).toBe(9)
})

test('9:01 should return next breakpoint (12), not current (9)', () => {
	const date = new Date()
	date.setHours(9)
	date.setMinutes(1)
	expect(calculateNearestHours(date)).toBe(12)
})
