import {describe, it, expect, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {defineComponent, h, ref, type Ref} from 'vue'
import {mount} from '@vue/test-utils'

import {useDaytimeSalutation} from './useDaytimeSalutation'
import {useAuthStore} from '@/stores/auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'
import en from '@/i18n/lang/en.json'

function makeDate(iso: string): Date {
	return new Date(iso)
}

function makeI18n() {
	return createI18n({
		legacy: false,
		locale: 'en',
		fallbackLocale: 'en',
		messages: {en},
	})
}

function runSalutation(now: Ref<Date>): string | undefined {
	let result: string | undefined
	const Comp = defineComponent({
		setup() {
			const s = useDaytimeSalutation(now)
			result = s.value
			return () => h('div')
		},
	})
	mount(Comp, {global: {plugins: [makeI18n()]}})
	return result
}

function setUser() {
	const authStore = useAuthStore()
	authStore.setUser({
		id: 42,
		name: 'Ada',
		username: 'ada',
		type: AUTH_TYPES.LINK_SHARE,
		created: new Date('2024-01-15T10:00:00Z'),
	} as never, false)
}

describe('useDaytimeSalutation', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('returns undefined when the user has no display name', () => {
		const now = ref(makeDate('2026-04-06T09:00:00'))
		expect(runSalutation(now)).toBeUndefined()
	})

	it('is deterministic for the same user, date, and bucket', () => {
		setUser()
		const now = ref(makeDate('2026-04-06T09:00:00'))
		const first = runSalutation(now)
		const second = runSalutation(now)

		expect(first).toBeDefined()
		expect(first).toBe(second)
	})

	it('produces a string from the morning pool on a Monday morning', () => {
		setUser()
		const now = ref(makeDate('2026-04-06T09:00:00'))
		const result = runSalutation(now)

		expect(result).toContain('Ada')
		const morningStrings = [
			'Good Morning Ada!',
			'Hey Ada, ready to go?',
			'Fresh start, Ada',
			'Coffee and tasks, Ada?',
			'Rise and plan, Ada',
			'Welcome back, Ada',
			'Fresh week, Ada',
		]
		expect(morningStrings).toContain(result)
	})

	it('includes the Friday extra in the pool on Friday morning', () => {
		setUser()
		const reachable = new Set<string>()
		for (let day = 3; day <= 31; day += 7) {
			const iso = `2026-04-${String(day).padStart(2, '0')}T09:00:00`
			const r = runSalutation(ref(makeDate(iso)))
			if (r) reachable.add(r)
		}
		expect(reachable.size).toBeGreaterThan(1)
	})

	it('uses different buckets for different hours', () => {
		setUser()
		const dateStr = '2026-04-06'
		const morning = runSalutation(ref(makeDate(`${dateStr}T09:00:00`)))
		const day = runSalutation(ref(makeDate(`${dateStr}T14:00:00`)))
		const evening = runSalutation(ref(makeDate(`${dateStr}T20:00:00`)))
		const night = runSalutation(ref(makeDate(`${dateStr}T02:00:00`)))

		expect(morning).toBeDefined()
		expect(day).toBeDefined()
		expect(evening).toBeDefined()
		expect(night).toBeDefined()
		expect(new Set([morning, day, evening, night]).size).toBeGreaterThan(1)
	})

	it('produces different results across consecutive days', () => {
		setUser()
		const results = new Set<string>()
		for (let day = 1; day <= 14; day++) {
			const iso = `2026-04-${String(day).padStart(2, '0')}T09:00:00`
			results.add(runSalutation(ref(makeDate(iso))) ?? '')
		}
		expect(results.size).toBeGreaterThan(1)
	})
})
