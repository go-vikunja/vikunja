import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {keepPreviousData, useQuery} from '@tanstack/vue-query'
import {activeTimerQuery, timeEntriesQuery} from '@/client/queries/timeEntries'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {PRO_FEATURE} from '@/constants/proFeatures'

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
			config.max_items_per_page,
		),
		enabled: toValue(enabled) && !auth.isLinkShareAuth && config.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING),
		...(keepPrevious ? {placeholderData: keepPreviousData} : {}),
	})))
	return {
		entries: computed(() => query.data.value?.items ?? []),
		totalPages: computed(() => query.data.value?.total_pages ?? 0),
	}
}
