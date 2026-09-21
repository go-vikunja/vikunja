import {queryOptions, useMutation} from '@tanstack/vue-query'
import {
	botsList,
	botsCreate,
	botsUpdate,
	botsDelete,
	type BotUser,
	type BotUserWritable,
	type BotUserReadBodyWritable,
} from '@/client/generated'
import {fetchAllPages} from './fetchAllPages'
import {contextMutationOptions} from './contextMutation'
import {apiTokenKeys} from './apiTokens'

export const botKeys = {all: ['bots'] as const}
export function botsQuery() {
	return queryOptions({
		queryKey: botKeys.all,
		queryFn: ({signal}) => fetchAllPages(async page => (await botsList({query: {page}, signal})).data),
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
			body: BotUserReadBodyWritable,
		}) => (await botsUpdate({path: {bot: id}, body})).data,
		onSuccess: (updated, {id}, client) => client.setQueryData<BotUser[]>(
			botKeys.all,
			current => current?.map(bot => bot.id === id ? updated : bot),
		),
		onSettled: (_input, client) => client.invalidateQueries({queryKey: botKeys.all}),
	})
}
export function deleteBotMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await botsDelete({path: {bot: id}})).data,
		onSuccess: (_data, id, client) => {
			client.setQueryData<BotUser[]>(botKeys.all, current => current?.filter(bot => bot.id !== id))
			client.removeQueries({queryKey: apiTokenKeys.list(id)})
		},
		onSettled: (_id, client) => client.invalidateQueries({queryKey: botKeys.all}),
	})
}
export function useCreateBotMutation() { return useMutation(createBotMutationOptions()) }
export function useUpdateBotMutation() { return useMutation(updateBotMutationOptions()) }
export function useDeleteBotMutation() { return useMutation(deleteBotMutationOptions()) }
