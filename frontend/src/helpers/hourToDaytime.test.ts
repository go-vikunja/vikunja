import {describe, it, expect} from 'vitest'
import {hourToDaytime} from './hourToDaytime'

function dateWithHour(hours: number): Date {
	const newDate = new Date()
	newDate.setHours(hours, 0, 0,0 )
	return newDate
}

describe('Salutation', () => {
	it('shows the right salutation in the night', () => {
		const salutation = hourToDaytime(dateWithHour(4))
		expect(salutation).toBe('night')
	})
	it('shows the right salutation in the morning', () => {
		const salutation = hourToDaytime(dateWithHour(8))
		expect(salutation).toBe('morning')
	})
	it('shows the right salutation in the day', () => {
		const salutation = hourToDaytime(dateWithHour(13))
		expect(salutation).toBe('day')
	})
	it('shows the right salutation in the night', () => {
		const salutation = hourToDaytime(dateWithHour(20))
		expect(salutation).toBe('evening')
	})
	it('shows the right salutation in the night again', () => {
		const salutation = hourToDaytime(dateWithHour(23))
		expect(salutation).toBe('night')
	})
})
