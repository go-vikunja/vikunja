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
	user: Pick<User, 'id' | 'name' | 'username'>,
}

// Chips render in key order, so a toggled emoji keeps its slot instead of moving to the end.
export function changeReaction(current: ReactionUsers = {}, input: ReactionInput) {
	const entries = Object.entries(current).map(([value, users]): [string, User[]] => [value, users ?? []])
	const existing = entries.find(([value]) => value === input.value)
	const remaining = (existing?.[1] ?? []).filter(user => user.id !== input.user.id)
	if (!input.remove) remaining.push(input.user)
	if (existing) existing[1] = remaining
	else if (remaining.length) entries.push([input.value, remaining])
	return Object.fromEntries(entries.filter(([value, users]) => value !== input.value || users.length > 0))
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
