import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQueries} from '@tanstack/vue-query'
import {linkSharesCreatedBy, linkSharesQuery} from '@/client/queries/linkShares'
import type {LinkSharing} from '@/client/generated'

export function useLinkSharesCreatedBy(projectIds: MaybeRefOrGetter<number[]>, userIds: MaybeRefOrGetter<number[]>) {
	const results = useQueries({
		queries: computed(() => {
			const users = toValue(userIds)
			if (users.length === 0) return []
			// select binds each result to its project, so ids can't misalign while the query list changes.
			return toValue(projectIds).map(projectId => ({
				...linkSharesQuery(projectId),
				select: (shares: LinkSharing[]) => linkSharesCreatedBy(projectId, shares, users),
			}))
		}),
	})
	return computed(() => results.value.flatMap(result => result.data ?? []))
}
