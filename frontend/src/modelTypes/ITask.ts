import type {Priority} from '@/constants/priorities'

import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ILabel} from './ILabel'
import type {IAttachment} from './IAttachment'
import type {ISubscription} from './ISubscription'
import type {IProject} from './IProject'
import type {IBucket} from './IBucket'

import type {IRelationKind} from '@/types/IRelationKind'
import type {IRepeatAfter} from '@/types/IRepeatAfter'
import type {IRepeatMode} from '@/types/IRepeatMode'

import type {PartialWithId} from '@/types/PartialWithId'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import type {IReactionPerEntity} from '@/modelTypes/IReaction'
import type {ITaskComment} from '@/modelTypes/ITaskComment.ts'

export interface ITask extends IAbstract {
	id: number
	title: string
	description: string
	done: boolean
	doneAt: Date | null
	priority: Priority
	labels: ILabel[]
	assignees: IUser[]

	dueDate: Date | null
	startDate: Date | null
	endDate: Date | null
	repeatAfter: number | IRepeatAfter
	repeatFromCurrentDate: boolean
	repeatMode: IRepeatMode
	reminders: ITaskReminder[]
	parentTaskId: ITask['id']
	hexColor: string
	percentDone: number
	relatedTasks: Partial<Record<IRelationKind, ITask[]>>
	attachments: IAttachment[]
	coverImageAttachmentId: IAttachment['id'] | null
	identifier: string
	index: number
	isFavorite: boolean
	subscription: ISubscription

	position: number
	
	reactions: IReactionPerEntity
	comments: ITaskComment[]

	createdBy: IUser
	created: Date
	updated: Date

	projectId: IProject['id'] // Meta, only used when creating a new task
	bucketId: IBucket['id']
}

export type ITaskPartialWithId = PartialWithId<ITask>
