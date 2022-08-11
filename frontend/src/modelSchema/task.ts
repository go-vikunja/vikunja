import type {ZodType, TypeOf} from 'zod'
import {nativeEnum, boolean, number, string, array, record, unknown, lazy} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'
import {HexColorSchema} from './common/hexColor'
import {TextFieldSchema} from './common/textField'
import {RelationKindSchema} from './common/RelationKind'
import {RepeatsSchema} from './common/repeats'

import {AbstractSchema} from './abstract'
import {AttachmentSchema} from './attachment'
import {LabelSchema} from './label'
import {SubscriptionSchema} from './subscription'
import {UserSchema} from './user'

import {PRIORITIES} from '@/constants/priorities'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'

const LabelsSchema = array(LabelSchema)
	.transform((labels) => labels.sort((f, s) => f.title > s.title ? 1 : -1)) // FIXME: use 
	.default([])

export type ILabels = TypeOf<typeof LabelsSchema> 

const RelatedTasksSchema = record(RelationKindSchema, record(string(), unknown()))
export type IRelatedTasksSchema = TypeOf<typeof RelatedTasksSchema>

const RelatedTasksLazySchema : ZodType<Task['relatedTasks']> = lazy(() => 
	record(RelationKindSchema, TaskSchema),
)
export type IRelatedTasksLazySchema = TypeOf<typeof RelatedTasksLazySchema>

// export interface ITask extends IAbstract {
// 	id: number
// 	title: string
// 	description: string
// 	done: boolean
// 	doneAt: Date | null
// 	priority: Priority
// 	labels: ILabel[]
// 	assignees: IUser[]

// 	dueDate: Date | null
// 	startDate: Date | null
// 	endDate: Date | null
// 	repeatAfter: number | IRepeatAfter
// 	repeatFromCurrentDate: boolean
// 	repeatMode: IRepeatMode
// 	reminderDates: Date[]
// 	parentTaskId: ITask['id']
// 	hexColor: string
// 	percentDone: number
// 	relatedTasks: Partial<Record<IRelationKind, ITask>>,
// 	attachments: IAttachment[]
// 	identifier: string
// 	index: number
// 	isFavorite: boolean
// 	subscription: ISubscription

// 	position: number
// 	kanbanPosition: number

// 	createdBy: IUser
// 	created: Date
// 	updated: Date

// 	listId: IList['id'] // Meta, only used when creating a new task
// 	bucketId: IBucket['id']
// }

export const TaskSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	title: TextFieldSchema, 
	description: TextFieldSchema,
	done: boolean().default(false),
	doneAt: DateSchema.nullable().default(null),
	priority: nativeEnum(PRIORITIES).default(PRIORITIES.UNSET),
	labels: LabelsSchema,
	assignees: array(UserSchema).default([]),

	dueDate: DateSchema.nullable(), // FIXME: default value is `0`. Shouldn't this be `null`?
	startDate: DateSchema.nullable(), // FIXME: default value is `0`. Shouldn't this be `null`?
	endDate: DateSchema.nullable(), // FIXME: default value is `0`. Shouldn't this be `null`?
	repeatAfter: RepeatsSchema, // FIXME: default value is `0`. Shouldn't this be `null`?
	repeatFromCurrentDate: boolean().default(false),
	repeatMode: nativeEnum(TASK_REPEAT_MODES).default(TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT),
	
	// TODO: schedule notifications
	// FIXME: triggered notificaitons not supported anymore / remove feature?
	reminderDates: array(DateSchema).default([]),
	parentTaskId: IdSchema.default(0), // shouldn't this have `null` as default?
	hexColor: HexColorSchema.default(''),
	percentDone: number().default(0),
	relatedTasks: RelatedTasksSchema.default({}),
	attachments: array(AttachmentSchema).default([]),
	identifier: string().default(''),
	index: number().default(0),
	isFavorite: boolean().default(false),
	subscription: SubscriptionSchema.nullable().default(null),

	position: number().default(0),
	kanbanPosition: number().default(0),

	createdBy: UserSchema,
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),

	listId: IdSchema.default(0), //IList['id'], // Meta, only used when creating a new task
	bucketId: IdSchema.default(0), // IBucket['id'],
}).transform((obj) => {
	if (obj.identifier === `-${obj.index}`) {
		obj.identifier = ''
	}
	return obj
})


export type Task = TypeOf<typeof TaskSchema> 
