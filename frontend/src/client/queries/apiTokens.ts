import {
	keepPreviousData,
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	tokensList,
	tokensCreate,
	tokensDelete,
	tokenRoutes,
	mcpInfo,
	type ApiToken,
	type ApiTokenWritable,
} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {
	removeFromPages,
	toPaginated,
	type Paginated,
} from './pagination'

export type ApiTokenPage = Paginated<ApiToken>

const apiTokenKeyRoot = ['apiTokens'] as const

export const apiTokenKeys = {
	all: apiTokenKeyRoot,
	lists: [...apiTokenKeyRoot, 'list'] as const,
	ownList: [...apiTokenKeyRoot, 'list', 'self'] as const,
	ownPage: (page: number) => [...apiTokenKeys.ownList, page] as const,
	list: (ownerId: number) => [...apiTokenKeys.lists, ownerId] as const,
	page: (ownerId: number, page: number) => [...apiTokenKeys.list(ownerId), page] as const,
	// Unpaged, so the MCP page can filter by permission client-side.
	ownAll: [...apiTokenKeyRoot, 'all'] as const,
	routes: [...apiTokenKeyRoot, 'routes'] as const,
	mcp: [...apiTokenKeyRoot, 'mcp'] as const,
}

export function apiTokensQuery(page: number) {
	return queryOptions({
		queryKey: apiTokenKeys.ownPage(page),
		queryFn: async ({signal}) => toPaginated((await tokensList({
			query: {page},
			signal,
		})).data, page),
		placeholderData: keepPreviousData,
	})
}

export function allApiTokensQuery() {
	return queryOptions({
		queryKey: apiTokenKeys.ownAll,
		queryFn: ({signal}) => fetchAllPages(async page => (await tokensList({
			query: {page},
			signal,
		})).data),
	})
}

export function botApiTokensQuery(ownerId: number, page: number) {
	return queryOptions({
		queryKey: apiTokenKeys.page(ownerId, page),
		queryFn: async ({signal}) => toPaginated((await tokensList({
			query: {
				page,
				owner_id: ownerId,
			},
			signal,
		})).data, page),
	})
}

export function apiTokenRoutesQuery() {
	return queryOptions({
		queryKey: apiTokenKeys.routes,
		queryFn: async ({signal}) => (await tokenRoutes({signal})).data,
		staleTime: Infinity,
	})
}

export function mcpInfoQuery() {
	return queryOptions({
		queryKey: apiTokenKeys.mcp,
		queryFn: async ({signal}) => (await mcpInfo({signal})).data,
	})
}

export function createApiTokenMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: ApiTokenWritable) => (await tokensCreate({body})).data,
		onSettled: (body, client) => body.owner_id === undefined
			? Promise.all([
				client.invalidateQueries({queryKey: apiTokenKeys.ownList}),
				client.invalidateQueries({queryKey: apiTokenKeys.ownAll}),
			])
			: client.invalidateQueries({queryKey: apiTokenKeys.list(body.owner_id)}),
	})
}

export function deleteApiTokenMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await tokensDelete({path: {id}})).data,
		onSuccess: (_data, id, client) => {
			client.setQueryData<ApiToken[]>(apiTokenKeys.ownAll, current => current?.filter(token => token.id !== id))
			removeFromPages<ApiToken>(client, apiTokenKeys.lists, token => token.id === id)
		},
		onSettled: (_id, client) => Promise.all([
			client.invalidateQueries({queryKey: apiTokenKeys.lists}),
			client.invalidateQueries({queryKey: apiTokenKeys.ownAll}),
		]),
	})
}

export function useCreateApiTokenMutation() {
	return useSecretMutation(createApiTokenMutationOptions())
}

export function useDeleteApiTokenMutation() {
	return useMutation(deleteApiTokenMutationOptions())
}
