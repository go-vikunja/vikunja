import {useMutation} from '@tanstack/vue-query'
import {
	subscriptionsCreate,
	subscriptionsDelete,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'
import {taskKeys} from './tasks'
import {i18n} from '@/i18n'

export function setTaskSubscriptionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, subscribed}: {
taskId: number,
subscribed: boolean,
}) => {
			const path = {
				entity: 'task',
				entityID: taskId,
			} as const
			if (!subscribed) {
				await subscriptionsDelete({path})
				return undefined
			}
			return (await subscriptionsCreate({path})).data
		},
		onSuccess: (subscription, {taskId}, client) => {
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				subscription,
			}))
		},
		onSettled: ({taskId}, client) => Promise.all([
			client.invalidateQueries({queryKey: taskKeys.detail(taskId)}),
			invalidateTaskMembership(client, taskId),
		]),
		successMessage: (_data, {subscribed}) => i18n.global.t(subscribed
			? 'task.subscription.subscribeSuccessTask' : 'task.subscription.unsubscribeSuccessTask'),
	})
}

export function useSetTaskSubscriptionMutation() {
	return useMutation(setTaskSubscriptionMutationOptions())
}
