import {mutationOptions, type QueryClient} from '@tanstack/vue-query'
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {error, success} from '@/message'

export function contextMutationOptions<TData, TInput>(options: {
	mutationFn: (input: TInput) => Promise<TData>
	onSuccess?: (data: TData, input: TInput, client: QueryClient) => void
	onSettled?: (input: TInput, client: QueryClient) => Promise<unknown>
	successMessage?: (data: TData, input: TInput) => string | undefined
}) {
	return mutationOptions({
		onMutate: () => captureClientRequestContext(),
		mutationFn: async (input: TInput) => {
			const request = captureClientRequestContext()
			const data = await options.mutationFn(input)
			assertClientRequestContext(request)
			return data
		},
		onSuccess: (data, input, request, {client}) => {
			assertClientRequestContext(request)
			options.onSuccess?.(data, input, client)
			const message = options.successMessage?.(data, input)
			if (message) success({message})
		},
		onError: (cause, _input, request) => {
			if (request && isClientRequestContextCurrent(request)) error(cause)
		},
		onSettled: async (_data, _cause, input, request, {client}) => {
			if (request && isClientRequestContextCurrent(request)) {
				await options.onSettled?.(input, client)
				assertClientRequestContext(request)
			}
		},
	})
}
