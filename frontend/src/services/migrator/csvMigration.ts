import AbstractService from '../abstractService'

export interface ColumnMapping {
	column_index: number
	column_name: string
	attribute: TaskAttribute
}

export type TaskAttribute =
	| 'title'
	| 'description'
	| 'due_date'
	| 'start_date'
	| 'end_date'
	| 'done'
	| 'priority'
	| 'labels'
	| 'project'
	| 'reminder'
	| 'ignore'

export const TASK_ATTRIBUTES: TaskAttribute[] = [
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

export interface DetectionResult {
	columns: string[]
	delimiter: string
	quote_char: string
	date_format: string
	suggested_mapping: ColumnMapping[]
	preview_rows: string[][]
}

export interface ImportConfig {
	delimiter: string
	quote_char: string
	date_format: string
	mapping: ColumnMapping[]
}

export interface PreviewTask {
	title: string
	description: string
	due_date?: string
	start_date?: string
	end_date?: string
	done: boolean
	priority: number
	labels?: string[]
	project?: string
}

export interface PreviewResult {
	tasks: PreviewTask[]
	total_rows: number
	error_count: number
	errors?: string[]
}

export interface MigrationStatus {
	started_at: string | null
	finished_at: string | null
}

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

export default class CSVMigrationService extends AbstractService {
	constructor() {
		super({})
	}

	getStatus(): Promise<MigrationStatus> {
		return this.getM('/migration/csv/status')
	}

	useCreateInterceptor() {
		return false
	}

	async detect(file: File): Promise<DetectionResult> {
		return this.uploadFile(
			'/migration/csv/detect',
			file,
			'import',
		)
	}

	async preview(file: File, config: ImportConfig): Promise<PreviewResult> {
		const data = new FormData()
		data.append('import', file)
		data.append('config', JSON.stringify(config))
		return this.uploadFormData('/migration/csv/preview', data)
	}

	async migrate(file: File, config: ImportConfig): Promise<{ message: string }> {
		const data = new FormData()
		data.append('import', file)
		data.append('config', JSON.stringify(config))
		return this.uploadFormData('/migration/csv/migrate', data)
	}
}
