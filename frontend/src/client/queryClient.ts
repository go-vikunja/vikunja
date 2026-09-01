import {QueryClient} from '@tanstack/vue-query'

export const queryClient = new QueryClient({
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
