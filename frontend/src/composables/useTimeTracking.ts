import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
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

export function useTimeEntries(filter: MaybeRefOrGetter<string>, enabled: MaybeRefOrGetter<boolean> = true) {
	const config = useConfigStore()
	const query = useQuery(computed(() => ({
		...timeEntriesQuery(toValue(filter), Intl.DateTimeFormat().resolvedOptions().timeZone),
		enabled: toValue(enabled) && config.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING),
	})))
	return {entries: computed(() => query.data.value ?? []), isPending: query.isPending}
}
