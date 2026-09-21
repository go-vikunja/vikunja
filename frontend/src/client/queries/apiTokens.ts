import {queryOptions, useMutation} from '@tanstack/vue-query'
import {tokensList, tokensCreate, tokensDelete, tokenRoutes, mcpInfo, type ApiToken, type ApiTokenWritable} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'

export const apiTokenKeys = {
	all: ['apiTokens'] as const,
	lists: ['apiTokens', 'list'] as const,
	list: (ownerId: number) => ['apiTokens', 'list', ownerId] as const,
	routes: ['apiTokens', 'routes'] as const,
	mcp: ['apiTokens', 'mcp'] as const,
}
export function apiTokensQuery(ownerId = 0) {
	return queryOptions({
		queryKey: apiTokenKeys.list(ownerId),
		queryFn: ({signal}) => fetchAllPages(async page => (await tokensList({query: {page, owner_id: ownerId || undefined}, signal})).data),
	})
}
export function apiTokenRoutesQuery() {
	return queryOptions({queryKey: apiTokenKeys.routes, queryFn: async ({signal}) => (await tokenRoutes({signal})).data})
}
export function mcpInfoQuery() {
	return queryOptions({queryKey: apiTokenKeys.mcp, queryFn: async ({signal}) => (await mcpInfo({signal})).data})
}
export function createApiTokenMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (body: ApiTokenWritable) => (await tokensCreate({body})).data,
			onSettled: (body, client) => client.invalidateQueries({queryKey: apiTokenKeys.list(body.owner_id ?? 0)}),
		}),
		// Response carries the plaintext token.
		gcTime: 0,
	}
}
export function deleteApiTokenMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await tokensDelete({path: {id}})).data,
		onSuccess: (_data, id, client) => client.setQueriesData<ApiToken[]>({queryKey: apiTokenKeys.lists}, current => current?.filter(token => token.id !== id)),
		onSettled: (_id, client) => client.invalidateQueries({queryKey: apiTokenKeys.lists}),
	})
}
export function useCreateApiTokenMutation() { return useMutation(createApiTokenMutationOptions()) }
export function useDeleteApiTokenMutation() { return useMutation(deleteApiTokenMutationOptions()) }
