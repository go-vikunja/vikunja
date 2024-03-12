import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export type ReactionKind = 'tasks' | 'comments'

export interface IReaction extends IAbstract {
	id: number
	kind: ReactionKind
	value: string
}

export interface IReactionPerEntity {
	[reaction: string]: IUser[]
}
