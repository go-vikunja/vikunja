import {computed, ref, toValue, type MaybeRefOrGetter} from 'vue'
import {useIntervalFn} from '@vueuse/core'
import {keepPreviousData, useQuery} from '@tanstack/vue-query'
import {activeTimerQuery, timeEntriesQuery} from '@/client/queries/timeEntries'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {PRO_FEATURE} from '@/constants/proFeatures'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export function useTimeTracking() {
	const auth = useAuthStore()
	const config = useConfigStore()
	const query = useQuery(computed(() => ({
		...activeTimerQuery(auth.info?.id ?? 0),
		enabled: !!auth.info?.id && !auth.isLinkShareAuth && config.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING),
	})))
	const activeTimer = computed(() => query.data.value ?? null)
	return {activeTimer, hasActiveTimer: computed(() => activeTimer.value !== null)}
}

// Live "m:ss" / "h:mm:ss" since startTime, ticking every second.
export function useElapsed(startTime: MaybeRefOrGetter<string | null | undefined>) {
	const now = ref(new Date())
	useIntervalFn(() => {
		now.value = new Date()
	}, 1000)
	return computed(() => {
		const start = parseDateOrNull(toValue(startTime))
		if (start === null) {
			return ''
		}
		const seconds = Math.max(0, Math.floor((now.value.getTime() - start.getTime()) / 1000))
		const pad = (n: number) => n.toString().padStart(2, '0')
		const hours = Math.floor(seconds / 3600)
		const mmss = `${pad(Math.floor((seconds % 3600) / 60))}:${pad(seconds % 60)}`
		return hours >= 1 ? `${hours}:${mmss}` : mmss
	})
}

export type UseTimeEntriesOptions = {
	enabled?: MaybeRefOrGetter<boolean>
	// Only for consumers whose filter stays inside one scope; a per-task list would show another task's entries.
	keepPrevious?: boolean
	page?: MaybeRefOrGetter<number>
}

export function useTimeEntries(
	filter: MaybeRefOrGetter<string>,
	{
		enabled = true,
		keepPrevious = false,
		page = 1,
	}: UseTimeEntriesOptions = {},
) {
	const auth = useAuthStore()
	const config = useConfigStore()
	const query = useQuery(computed(() => ({
		...timeEntriesQuery(
			toValue(filter),
			Intl.DateTimeFormat().resolvedOptions().timeZone,
			toValue(page),
		),
		enabled: toValue(enabled) && !auth.isLinkShareAuth && config.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING),
		...(keepPrevious ? {placeholderData: keepPreviousData} : {}),
	})))
	return {
		entries: computed(() => query.data.value?.items ?? []),
		totalPages: computed(() => query.data.value?.total_pages ?? 0),
	}
}
