import {CancelledError, QueryCache, QueryClient} from '@tanstack/vue-query'

import {isRequestContextAbort} from '@/client/requestContext'
import {isInTrashError} from '@/client/queries/trashShared'
import {error} from '@/message'

export const queryClient = new QueryClient({
	queryCache: new QueryCache({
		onError: (cause, query) => {
			if (query.meta?.handlesError) return
			if (cause instanceof CancelledError || isRequestContextAbort(cause)) return
			if (isInTrashError(cause)) return
			error(cause)
		},
	}),
	defaultOptions: {
		queries: {
			staleTime: 60_000,
			retry: 1,
			refetchOnWindowFocus: false,
		},
		mutations: {
			retry: false,
		},
	},
})
