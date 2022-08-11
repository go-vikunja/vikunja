import type {TypeOf} from 'zod'
import {nativeEnum} from 'zod'

export const RELATION_KIND = {
	'SUBTASK': 'subtask',
	'PARENTTASK': 'parenttask',
	'RELATED': 'related',
	'DUPLICATES': 'duplicates',
	'BLOCKING': 'blocking',
	'BLOCKED': 'blocked',
	'PROCEDES': 'precedes',
	'FOLLOWS': 'follows',
	'COPIEDFROM': 'copiedfrom',
	'COPIEDTO': 'copiedto',
} as const

export const RELATION_KINDS = [...Object.values(RELATION_KIND)] as const

export const RelationKindSchema = nativeEnum(RELATION_KIND)

export type IRelationKind = TypeOf<typeof RelationKindSchema>
 