export enum RELATION_KIND {
	'SUBTASK' = 'subtask',
	'PARENTTASK' = 'parenttask',
	'RELATED' = 'related',
	'DUPLICATES' = 'duplicates',
	'BLOCKING' = 'blocking',
	'BLOCKED' = 'blocked',
	'PROCEDES' = 'precedes',
	'FOLLOWS' = 'follows',
	'COPIEDFROM' = 'copiedfrom',
	'COPIEDTO' = 'copiedto',
}
	
export type IRelationKind = typeof RELATION_KIND[keyof typeof RELATION_KIND] 
 
export const RELATION_KINDS = [...Object.values(RELATION_KIND)] as const
