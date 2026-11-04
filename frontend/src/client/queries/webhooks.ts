import {
	keepPreviousData,
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	webhooksList,
	webhooksCreate,
	webhooksUpdate,
	webhooksDelete,
	webhooksEventsList,
	userWebhooksList,
	userWebhooksCreate,
	userWebhooksUpdate,
	userWebhooksDelete,
	userWebhooksEvents,
} from '@/client/generated'
import type {
	Webhook,
	WebhookWritable,
	WebhooksUpdateData,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {
	removeFromPages,
	toPaginated,
	type Paginated,
} from './pagination'
import {i18n} from '@/i18n'

export type WebhookScope = {kind: 'project', projectId: number} | {kind: 'user'}
export type WebhookPage = Paginated<Webhook>
export const webhookKeys = {
	projects: ['webhooks', 'list', 'project'] as const,
	scope: (scope: WebhookScope) => scope.kind === 'project'
		? ['webhooks', 'list', 'project', scope.projectId] as const
		: ['webhooks', 'list', 'user'] as const,
	list: (scope: WebhookScope, page: number) => [...webhookKeys.scope(scope), page] as const,
	events: (kind: WebhookScope['kind']) => ['webhooks', 'events', kind] as const,
}
export function webhooksQuery(scope: WebhookScope, page: number) {
	return queryOptions({
		queryKey: webhookKeys.list(scope, page),
		queryFn: async ({signal}): Promise<WebhookPage> => {
			const query = {page}
			const {data} = await (scope.kind === 'project'
				? webhooksList({
					path: {project: scope.projectId},
					query,
					signal,
				})
				: userWebhooksList({
					query,
					signal,
				}))
			return toPaginated(data, page)
		},
		placeholderData: keepPreviousData,
	})
}
export function webhookEventsQuery(kind: WebhookScope['kind']) {
	return queryOptions({
		queryKey: webhookKeys.events(kind),
		queryFn: async ({signal}) => (await (kind === 'project'
			? webhooksEventsList({signal})
			: userWebhooksEvents({signal}))).data ?? [],
	})
}
export function createWebhookMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, body}: {scope: WebhookScope, body: WebhookWritable}) => (await (scope.kind === 'project'
			? webhooksCreate({path: {project: scope.projectId}, body})
			: userWebhooksCreate({body}))).data,
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.scope(scope)}),
	})
}
export function deleteWebhookMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, id}: {scope: WebhookScope, id: number}) => (await (scope.kind === 'project'
			? webhooksDelete({path: {project: scope.projectId, webhook: id}})
			: userWebhooksDelete({path: {webhook: id}}))).data,
		onSuccess: (_data, {scope, id}, client) => removeFromPages<Webhook>(client, webhookKeys.scope(scope), webhook => webhook.id === id),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.scope(scope)}),
		successMessage: () => i18n.global.t('project.webhooks.deleteSuccess'),
	})
}
export function updateWebhookMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, id, body}: {scope: WebhookScope, id: number, body: WebhooksUpdateData['body']}) => (await (scope.kind === 'project'
			? webhooksUpdate({path: {project: scope.projectId, webhook: id}, body})
			: userWebhooksUpdate({path: {webhook: id}, body}))).data,
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.scope(scope)}),
	})
}
export function useCreateWebhookMutation() { return useSecretMutation(createWebhookMutationOptions()) }
export function useDeleteWebhookMutation() { return useMutation(deleteWebhookMutationOptions()) }
export function useUpdateWebhookMutation() { return useSecretMutation(updateWebhookMutationOptions()) }
