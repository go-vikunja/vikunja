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
	taskKeys,
	type TaskExpansion,
	type TaskResponse,
} from './tasks'

export type ReactionKind = ReactionsCreateData['path']['entitykind']
export type ReactionUsers = NonNullable<Task['reactions']>
type ReactionInput = {
	kind: ReactionKind,
	id: number,
	value: string,
	remove: boolean,
	user: Pick<User, 'id' | 'name' | 'username'>,
}

// taskKeys.detail(id, expand) puts the expansion last.
const DETAIL_KEY_EXPANSION = 3

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
			if (input.kind !== 'tasks') return
			const reacted = {
				...input,
				user: data?.user ?? input.user,
			}
			// Reactions are expand-only, so patching a copy that never asked for them fakes a complete map.
			for (const [key, cached] of client.getQueriesData<TaskResponse>({queryKey: taskKeys.details})) {
				const expand = key[DETAIL_KEY_EXPANSION] as TaskExpansion | undefined
				if (cached?.id !== input.id || !expand?.includes('reactions')) continue
				client.setQueryData<TaskResponse>(key, {
					...cached,
					reactions: changeReaction(cached.reactions, reacted),
				})
			}
		},
	})
}

export function useSetReactionMutation() {
	return useMutation(setReactionMutationOptions())
}
