import {queryOptions, useMutation} from '@tanstack/vue-query'
import {webhooksList, webhooksCreate, webhooksDelete, webhooksEventsList, userWebhooksList, userWebhooksCreate, userWebhooksDelete, userWebhooksEvents, type Webhook, type WebhookWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {i18n} from '@/i18n'

export type WebhookScope = {kind: 'project', projectId: number} | {kind: 'user'}
export const webhookKeys = {
	list: (scope: WebhookScope) => scope.kind === 'project'
		? ['webhooks', 'list', 'project', scope.projectId] as const
		: ['webhooks', 'list', 'user'] as const,
	events: (kind: WebhookScope['kind']) => ['webhooks', 'events', kind] as const,
}
export function webhooksQuery(scope: WebhookScope) {
	return queryOptions({
		queryKey: webhookKeys.list(scope),
		queryFn: ({signal}) => fetchAllPages(async page => (await (scope.kind === 'project'
			? webhooksList({path: {project: scope.projectId}, query: {page}, signal})
			: userWebhooksList({query: {page}, signal}))).data),
	})
}
export function webhookEventsQuery(kind: WebhookScope['kind']) {
	return queryOptions({
		queryKey: webhookKeys.events(kind),
		queryFn: async ({signal}) => (await (kind === 'project' ? webhooksEventsList({signal}) : userWebhooksEvents({signal}))).data ?? [],
	})
}
export function createWebhookMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, body}: {scope: WebhookScope, body: WebhookWritable}) => (await (scope.kind === 'project'
			? webhooksCreate({path: {project: scope.projectId}, body}) : userWebhooksCreate({body}))).data,
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.list(scope)}),
	})
}
export function deleteWebhookMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, id}: {scope: WebhookScope, id: number}) => (await (scope.kind === 'project'
			? webhooksDelete({path: {project: scope.projectId, webhook: id}}) : userWebhooksDelete({path: {webhook: id}}))).data,
		onSuccess: (_data, {scope, id}, client) => client.setQueryData<Webhook[]>(webhookKeys.list(scope), current => current?.filter(webhook => webhook.id !== id)),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.list(scope)}),
		successMessage: () => i18n.global.t('project.webhooks.deleteSuccess'),
	})
}
export function useCreateWebhookMutation() { return useMutation(createWebhookMutationOptions()) }
export function useDeleteWebhookMutation() { return useMutation(deleteWebhookMutationOptions()) }
