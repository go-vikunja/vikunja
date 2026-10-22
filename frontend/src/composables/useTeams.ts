import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {teamQuery, teamSearchQuery, teamsPageQuery} from '@/client/queries/teams'

export function useTeamsPage(page: MaybeRefOrGetter<number>) {
	const query = useQuery(computed(() => teamsPageQuery(toValue(page))))
	return {
		...query,
		teams: computed(() => query.data.value?.items ?? []),
		totalPages: computed(() => query.data.value?.total_pages ?? 0),
	}
}

export function useTeamSearch({search, includePublic = false, enabled = true}: {
	search: MaybeRefOrGetter<string>
	includePublic?: MaybeRefOrGetter<boolean>
	enabled?: MaybeRefOrGetter<boolean>
}) {
	const query = useQuery(computed(() => ({...teamSearchQuery(toValue(search), toValue(includePublic)), enabled: toValue(enabled)})))
	return {...query, teams: computed(() => query.data.value ?? [])}
}

export function useTeam(id: MaybeRefOrGetter<number>) {
	return useQuery(computed(() => ({...teamQuery(toValue(id)), enabled: toValue(id) > 0})))
}
