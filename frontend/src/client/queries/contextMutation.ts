import {mutationOptions, type QueryClient, type QueryKey} from '@tanstack/vue-query'
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {error, success} from '@/message'

export function contextMutationOptions<TData, TInput, TOptimistic = undefined>(options: {
	mutationFn: (input: TInput) => Promise<TData>
	optimistic?: {
		queryKeys: (input: TInput, client: QueryClient) => readonly QueryKey[]
		update: (input: TInput, client: QueryClient) => TOptimistic
	}
	onSuccess?: (data: TData, input: TInput, client: QueryClient, optimistic: TOptimistic) => void
	onSettled?: (input: TInput, client: QueryClient) => Promise<unknown>
	successMessage?: (data: TData, input: TInput) => string | undefined
	toastError?: (input: TInput) => boolean
}) {
	return mutationOptions({
		onMutate: async (input: TInput, {client}) => {
			const request = captureClientRequestContext()
			if (!options.optimistic) {
				return {request, previous: [], optimistic: undefined as TOptimistic}
			}
			const queryKeys = options.optimistic.queryKeys(input, client)
			await Promise.all(queryKeys.map(queryKey => client.cancelQueries({queryKey})))
			assertClientRequestContext(request)
			const previous = queryKeys.flatMap(queryKey => client.getQueriesData({queryKey}))
			return {request, previous, optimistic: options.optimistic.update(input, client)}
		},
		mutationFn: async (input: TInput) => {
			const request = captureClientRequestContext()
			const data = await options.mutationFn(input)
			assertClientRequestContext(request)
			return data
		},
		onSuccess: (data, input, context, {client}) => {
			assertClientRequestContext(context.request)
			options.onSuccess?.(data, input, client, context.optimistic)
			const message = options.successMessage?.(data, input)
			if (message) success({message})
		},
		onError: (cause, input, context, {client}) => {
			if (!context || !isClientRequestContextCurrent(context.request)) return
			for (const [queryKey, previous] of context.previous) {
				if (previous !== undefined) client.setQueryData(queryKey, previous)
			}
			if (options.toastError?.(input) ?? true) error(cause)
		},
		onSettled: async (_data, _cause, input, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await options.onSettled?.(input, client)
				assertClientRequestContext(context.request)
			}
		},
	})
}
