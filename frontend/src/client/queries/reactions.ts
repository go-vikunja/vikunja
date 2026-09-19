import {useMutation} from '@tanstack/vue-query'
import {
	reactionsCreate,
	reactionsDelete,
} from '@/client/generated'
import type {
	ReactionsCreateData,
	Task,
	User,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'

export type ReactionKind = ReactionsCreateData['path']['entitykind']
export type ReactionUsers = NonNullable<Task['reactions']>
type ReactionInput = {
	kind: ReactionKind,
	id: number,
	value: string,
	remove: boolean,
	user: Pick<User, 'id' | 'name' | 'username' | 'bot_owner_id'>,
}

export function changeReaction(current: ReactionUsers = {}, input: ReactionInput) {
	const {[input.value]: users, ...rest} = Object.fromEntries(
		Object.entries(current).map(([value, users]) => [value, users ?? []]),
	)
	const remaining = (users ?? []).filter(user => user.id !== input.user.id)
	if (!input.remove) remaining.push(input.user)
	return remaining.length ? {
		...rest,
		[input.value]: remaining,
	} : rest
}

export function setReactionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (input: ReactionInput) => {
			const request = {
				path: {
					entitykind: input.kind,
					entityid: input.id,
				},
				body: {value: input.value},
			}
			if (input.remove) {
				await reactionsDelete(request)
				return undefined
			}
			return (await reactionsCreate(request)).data
		},
		onSuccess: (data, input, client) => {
			if (input.kind === 'tasks') {
				const reacted = {
					...input,
					user: data?.user ?? input.user,
				}
				mapTaskEverywhere(client, input.id, task => ({
					...task,
					reactions: changeReaction(task.reactions, reacted),
				}))
			}
		},
		onSettled: (input, client) => input.kind === 'tasks'
			? invalidateTaskMembership(client, input.id)
			: Promise.resolve(),
	})
}

export function useSetReactionMutation() {
	return useMutation(setReactionMutationOptions())
}
