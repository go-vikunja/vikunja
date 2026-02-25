import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'

export interface IAutoTaskTemplate {
	id?: number
	project_id: number
	title: string
	description: string
	priority: number
	hex_color: string
	label_ids: number[]
	assignee_ids: number[]
	interval_value: number
	interval_unit: string // 'hours' | 'days' | 'weeks' | 'months'
	start_date: string
	end_date: string | null
	active: boolean
	last_created_at?: string | null
	next_due_at?: string | null
	owner?: any
	created?: string
	updated?: string
}

export function emptyAutoTaskTemplate(): IAutoTaskTemplate {
	return {
		project_id: 0,
		title: '',
		description: '',
		priority: 0,
		hex_color: '',
		label_ids: [],
		assignee_ids: [],
		interval_value: 1,
		interval_unit: 'days',
		start_date: new Date().toISOString(),
		end_date: null,
		active: true,
	}
}

const http = AuthenticatedHTTPFactory()

export async function getAllAutoTasks(): Promise<IAutoTaskTemplate[]> {
	const response = await http.get('/autotasks')
	return response.data
}

export async function getAutoTask(id: number): Promise<IAutoTaskTemplate> {
	const response = await http.get(`/autotasks/${id}`)
	return response.data
}

export async function createAutoTask(data: IAutoTaskTemplate): Promise<IAutoTaskTemplate> {
	const response = await http.put('/autotasks', data)
	return response.data
}

export async function updateAutoTask(data: IAutoTaskTemplate): Promise<IAutoTaskTemplate> {
	const response = await http.post(`/autotasks/${data.id}`, data)
	return response.data
}

export async function deleteAutoTask(id: number): Promise<void> {
	await http.delete(`/autotasks/${id}`)
}

export async function triggerAutoTask(id: number): Promise<any> {
	const response = await http.post(`/autotasks/${id}/trigger`, {})
	return response.data
}

export async function checkAutoTasks(): Promise<{created: any[]}> {
	const response = await http.post('/autotasks/check', {})
	return response.data
}

export async function truncateAutoTaskLog(templateId: number, keep: number = 0): Promise<{deleted: number}> {
	const response = await http.post(`/autotasks/${templateId}/log/truncate?keep=${keep}`, {})
	return response.data
}
