import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {teamQuery, teamsQuery} from '@/client/queries/teams'

export function useTeams({search = '', includePublic = false, enabled = true}: {
	search?: MaybeRefOrGetter<string>
	includePublic?: MaybeRefOrGetter<boolean>
	enabled?: MaybeRefOrGetter<boolean>
} = {}) {
	const query = useQuery(computed(() => ({...teamsQuery(toValue(search), toValue(includePublic)), enabled: toValue(enabled)})))
	return {...query, teams: computed(() => query.data.value ?? [])}
}

export function useTeam(id: MaybeRefOrGetter<number>) {
	return useQuery(computed(() => ({...teamQuery(toValue(id)), enabled: toValue(id) > 0})))
}
