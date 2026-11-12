import {keepPreviousData, queryOptions, useMutation} from '@tanstack/vue-query'
import {
	botsList,
	botsCreate,
	botsUpdate,
	botsDelete,
	type BotUser,
	type BotUserWritable,
	type BotUserUpdateBodyWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {apiTokenKeys} from './apiTokens'
import {
	removeFromPages,
	toPaginated,
	type Paginated,
} from './pagination'

export type BotPage = Paginated<BotUser>
export const botKeys = {
	all: ['bots'] as const,
	list: (page: number) => ['bots', 'list', page] as const,
}
export function botsQuery(page: number) {
	return queryOptions({
		queryKey: botKeys.list(page),
		queryFn: async ({signal}): Promise<BotPage> => {
			const {data} = await botsList({
				query: {page},
				signal,
			})
			return toPaginated(data, page)
		},
		placeholderData: keepPreviousData,
	})
}
export function createBotMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: BotUserWritable) => (await botsCreate({body})).data,
		onSettled: (_body, client) => client.invalidateQueries({queryKey: botKeys.all}),
		toastError: () => false,
	})
}
export function updateBotMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, body}: {
			id: number,
			body: BotUserUpdateBodyWritable,
		}) => (await botsUpdate({path: {bot: id}, body})).data,
		onSuccess: (updated, {id}, client) => client.setQueriesData<BotPage>(
			{queryKey: botKeys.all},
			current => current && {
				...current,
				items: current.items.map(bot => bot.id === id ? updated : bot),
			},
		),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: botKeys.all}),
	})
}
export function deleteBotMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await botsDelete({path: {bot: id}})).data,
		onSuccess: (_data, id, client) => {
			removeFromPages<BotUser>(client, botKeys.all, bot => bot.id === id)
			client.removeQueries({queryKey: apiTokenKeys.list(id)})
		},
		onSettled: (_id, client) => client.invalidateQueries({queryKey: botKeys.all}),
	})
}
export function useCreateBotMutation() { return useMutation(createBotMutationOptions()) }
export function useUpdateBotMutation() { return useMutation(updateBotMutationOptions()) }
export function useDeleteBotMutation() { return useMutation(deleteBotMutationOptions()) }
