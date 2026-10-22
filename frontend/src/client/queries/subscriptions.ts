import {useMutation} from '@tanstack/vue-query'
import {contextMutationOptions} from './contextMutation'
import {setSubscription} from './subscriptionRequests'
import {mapTaskEverywhere} from './taskCache'
import {taskKeys} from './tasks'
import {i18n} from '@/i18n'

export function setTaskSubscriptionMutationOptions() {
	return contextMutationOptions({
		mutationFn: ({taskId, subscribed}: {
			taskId: number,
			subscribed: boolean,
		}) => setSubscription('task', taskId, subscribed),
		onSuccess: (subscription, {taskId}, client) => {
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				subscription,
			}))
		},
		onSettled: ({taskId}, client) => client.invalidateQueries({queryKey: [...taskKeys.details, taskId]}),
		successMessage: (_data, {subscribed}) => i18n.global.t(subscribed
			? 'task.subscription.subscribeSuccessTask'
			: 'task.subscription.unsubscribeSuccessTask'),
	})
}

export function useSetTaskSubscriptionMutation() {
	return useMutation(setTaskSubscriptionMutationOptions())
}
