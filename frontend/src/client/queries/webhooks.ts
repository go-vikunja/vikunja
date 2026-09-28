import {
	keepPreviousData,
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	webhooksList,
	webhooksCreate,
	webhooksDelete,
	webhooksEventsList,
	userWebhooksList,
	userWebhooksCreate,
	userWebhooksDelete,
	userWebhooksEvents,
} from '@/client/generated'
import type {
	Webhook,
	WebhookWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
import {
	totalPagesFor,
	type Paginated,
} from './pagination'
import {i18n} from '@/i18n'

export type WebhookScope = {kind: 'project', projectId: number} | {kind: 'user'}
export type WebhookPage = Paginated<Webhook>
export const webhookKeys = {
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
			return {
				items: data.items ?? [],
				page: data.page ?? page,
				per_page: data.per_page ?? 0,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
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
		onSuccess: (_data, {scope, id}, client) => client.setQueriesData<WebhookPage>(
			{queryKey: webhookKeys.scope(scope)},
			current => {
				if (!current) return current
				const total = Math.max(0, current.total - 1)
				return {
					...current,
					items: current.items.filter(webhook => webhook.id !== id),
					total,
					total_pages: totalPagesFor(current, total),
				}
			},
		),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: webhookKeys.scope(scope)}),
		successMessage: () => i18n.global.t('project.webhooks.deleteSuccess'),
	})
}
export function useCreateWebhookMutation() { return useSecretMutation(createWebhookMutationOptions()) }
export function useDeleteWebhookMutation() { return useMutation(deleteWebhookMutationOptions()) }
