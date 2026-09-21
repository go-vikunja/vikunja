import type {ColumnMapping} from '@/client/generated'

export type CsvImportDraft = {delimiter: string, quote_char: string, date_format: string, skip_rows: number, mapping: ColumnMapping[]}

export const TASK_ATTRIBUTES: NonNullable<ColumnMapping['attribute']>[] = [
	'title',
	'description',
	'due_date',
	'start_date',
	'end_date',
	'done',
	'priority',
	'labels',
	'project',
	'reminder',
	'ignore',
]

export const SUPPORTED_DELIMITERS = [',', ';', '\t', '|'] as const

export const SUPPORTED_DATE_FORMATS = [
	'2006-01-02',
	'2006-01-02T15:04:05',
	'02/01/2006',
	'01/02/2006',
	'02-01-2006',
	'01-02-2006',
	'02.01.2006',
	'2006/01/02',
	'2006-01-02 15:04:05',
] as const

