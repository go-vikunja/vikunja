import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useNow} from '@vueuse/core'

import {useAuthStore} from '@/stores/auth'

type Daytime = 'night' | 'morning' | 'day' | 'evening'

export function hourToDaytime(now: Date): Daytime {
	const hours = now.getHours()

	const daytimeMap = {
		night: hours < 5 || hours > 23,
		morning: hours < 11,
		day: hours < 18,
		evening: hours < 23,
	} as Record<Daytime, boolean>

	return (Object.keys(daytimeMap) as Daytime[]).find((daytime) => daytimeMap[daytime]) || 'night'
}

export function useDateTimeSalutation() {
	const {t} = useI18n({useScope: 'global'})
	const now = useNow()
	const authStore = useAuthStore()

	const name = computed(() => authStore.userDisplayName)
	const daytime = computed(() => hourToDaytime(now.value))

	const salutations = {
		'night': () => t('home.welcomeNight', {username: name.value}),
		'morning': () => t('home.welcomeMorning', {username: name.value}),
		'day': () => t('home.welcomeDay', {username: name.value}),
		'evening': () => t('home.welcomeEvening', {username: name.value}),
	} as Record<Daytime, () => string>

	return computed(() => name.value ? salutations[daytime.value]() : undefined)
}