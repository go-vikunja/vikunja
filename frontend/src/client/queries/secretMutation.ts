import {
	useMutation,
	type DefaultError,
	type MutationOptions,
	type UseMutationReturnType,
} from '@tanstack/vue-query'

// gcTime 0 only evicts once no observer holds the mutation, so every settle also resets it.
export function useSecretMutation<TData = unknown, TError = DefaultError, TVariables = void, TContext = unknown>(
	options: MutationOptions<TData, TError, TVariables, TContext>,
): UseMutationReturnType<TData, TError, TVariables, TContext> {
	const mutation = useMutation({
		...options,
		gcTime: 0,
	})
	const mutateAsync: typeof mutation.mutateAsync = async (...args) => {
		try {
			return await mutation.mutateAsync(...args)
		} finally {
			// A newer call may already be pending on this observer; resetting would detach it.
			if (!mutation.isPending.value) {
				mutation.reset()
			}
		}
	}
	return {
		...mutation,
		mutateAsync,
		mutate: (...args) => {
			mutateAsync(...args).catch(() => {})
		},
	}
}
