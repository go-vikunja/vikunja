import {computed, onActivated, ref, type Ref} from 'vue'
import {useI18n} from 'vue-i18n'

import {useAuthStore} from '@/stores/auth'
import {hourToDaytime} from '@/helpers/hourToDaytime'
import {stringHash} from '@/helpers/stringHash'

export type Daytime = 'night' | 'morning' | 'day' | 'evening'

// Base i18n keys for each bucket. Existing keys (welcomeNight/Morning/Day/Evening)
// are kept as the first entry of their respective pool so prior translations remain valid.
const basePools: Record<Daytime, string[]> = {
	night: [
		'home.welcomeNight',
		'home.welcomeNightOwl',
		'home.welcomeNightBurning',
		'home.welcomeNightQuiet',
		'home.welcomeNightLate',
		'home.welcomeNightMoonlit',
	],
	morning: [
		'home.welcomeMorning',
		'home.welcomeMorningHey',
		'home.welcomeMorningFresh',
		'home.welcomeMorningCoffee',
		'home.welcomeMorningRise',
		'home.welcomeMorningBack',
	],
	day: [
		'home.welcomeDay',
		'home.welcomeDayBack',
		'home.welcomeDayFocus',
		'home.welcomeDayKeepGoing',
		'home.welcomeDayWhatsNext',
		'home.welcomeDayGood',
	],
	evening: [
		'home.welcomeEvening',
		'home.welcomeEveningWind',
		'home.welcomeEveningReturns',
		'home.welcomeEveningWrap',
		'home.welcomeEveningOneMore',
		'home.welcomeEveningStill',
	],
}

// One entry per weekday (index = Date.getDay(), Sunday = 0). Appended to the
// morning pool only, on its matching day.
const morningWeekdayExtras: (string | null)[] = [
	'home.welcomeSundaySession', // 0 Sun
	'home.welcomeMondayFresh',   // 1 Mon
	'home.welcomeTuesday',       // 2 Tue
	'home.welcomeWednesdayMid',  // 3 Wed
	'home.welcomeThursday',      // 4 Thu
	'home.welcomeFridayPush',    // 5 Fri
	'home.welcomeSaturday',      // 6 Sat
]

function poolFor(bucket: Daytime, now: Date): string[] {
	if (bucket !== 'morning') {
		return basePools[bucket]
	}
	const extra = morningWeekdayExtras[now.getDay()]
	return extra ? [...basePools.morning, extra] : basePools.morning
}

function dateKey(now: Date): string {
	return `${now.getFullYear()}-${now.getMonth() + 1}-${now.getDate()}`
}

export function useDaytimeSalutation(now?: Ref<Date>) {
	const {t} = useI18n({useScope: 'global'})
	const internalNow = ref(new Date())
	const currentDate = now ?? internalNow
	onActivated(() => {
		internalNow.value = new Date()
	})
	const authStore = useAuthStore()

	const name = computed(() => authStore.userDisplayName)
	// Use the user's created timestamp as the per-user hash component.
	// It's stable, unique per user, and doesn't leak the sequential user id.
	const userKey = computed(() => authStore.info?.created?.getTime() ?? 0)
	const bucket = computed(() => hourToDaytime(currentDate.value))

	return computed(() => {
		if (!name.value) {
			return undefined
		}
		const pool = poolFor(bucket.value, currentDate.value)
		const key = `${dateKey(currentDate.value)}_${bucket.value}_${userKey.value}`
		const index = stringHash(key) % pool.length
		return t(pool[index], {username: name.value})
	})
}
