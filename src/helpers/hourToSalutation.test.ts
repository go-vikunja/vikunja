import {describe, it, expect} from 'vitest'

import {hourToSalutation} from './hourToSalutation'

const dateWithHour = (hours: number): Date => {
	const date = new Date()
	date.setHours(hours)
	return date
}

describe('Salutation', () => {
	it('shows the right salutation in the night', () => {
		const salutation = hourToSalutation(dateWithHour(4))
		expect(salutation).toBe('home.welcomeNight')
	})
	it('shows the right salutation in the morning', () => {
		const salutation = hourToSalutation(dateWithHour(8))
		expect(salutation).toBe('home.welcomeMorning')
	})
	it('shows the right salutation in the day', () => {
		const salutation = hourToSalutation(dateWithHour(13))
		expect(salutation).toBe('home.welcomeDay')
	})
	it('shows the right salutation in the night', () => {
		const salutation = hourToSalutation(dateWithHour(20))
		expect(salutation).toBe('home.welcomeEvening')
	})
	it('shows the right salutation in the night again', () => {
		const salutation = hourToSalutation(dateWithHour(23))
		expect(salutation).toBe('home.welcomeNight')
	})
})


