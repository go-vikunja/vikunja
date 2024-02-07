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
