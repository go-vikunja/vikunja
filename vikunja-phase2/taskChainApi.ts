import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'

export interface ITaskChainStepAttachment {
	id: number
	step_id: number
	file_name: string
	file?: {
		id: number
		name: string
		size: number
		mime: string
	}
	created?: string
}

export type TimeUnit = 'hours' | 'days' | 'weeks' | 'months'

export interface ITaskChainStep {
	id?: number
	chain_id?: number
	sequence: number
	title: string
	description: string
	offset_days: number
	offset_unit?: TimeUnit
	duration_days: number
	duration_unit?: TimeUnit
	priority: number
	hex_color: string
	label_ids: number[]
	attachments?: ITaskChainStepAttachment[]
}

/** Convert a value in the given unit to fractional days */
export function unitToDays(value: number, unit: TimeUnit | undefined): number {
	switch (unit) {
		case 'hours': return value / 24
		case 'weeks': return value * 7
		case 'months': return value * 30
		default: return value
	}
}

/** Convert fractional days to the display value for a given unit */
export function daysToUnit(days: number, unit: TimeUnit | undefined): number {
	switch (unit) {
		case 'hours': return Math.round(days * 24)
		case 'weeks': return +(days / 7).toFixed(1)
		case 'months': return +(days / 30).toFixed(1)
		default: return days
	}
}

export interface ITaskChain {
	id?: number
	title: string
	description: string
	steps: ITaskChainStep[]
	owner?: any
	created?: string
	updated?: string
}

export interface ITaskFromChainRequest {
	target_project_id: number
	anchor_date: string
	title_prefix?: string
	custom_steps?: ITaskChainStep[]
}

const http = AuthenticatedHTTPFactory()

export async function getAllChains(): Promise<ITaskChain[]> {
	const response = await http.get('/taskchains')
	return response.data
}

export async function getChain(id: number): Promise<ITaskChain> {
	const response = await http.get(`/taskchains/${id}`)
	return response.data
}

export async function createChain(chain: ITaskChain): Promise<ITaskChain> {
	const response = await http.put('/taskchains', chain)
	return response.data
}

export async function updateChain(chain: ITaskChain): Promise<ITaskChain> {
	const response = await http.post(`/taskchains/${chain.id}`, chain)
	return response.data
}

export async function deleteChain(id: number): Promise<void> {
	await http.delete(`/taskchains/${id}`)
}

export async function createTasksFromChain(chainId: number, request: ITaskFromChainRequest): Promise<any> {
	const response = await http.put(`/taskchains/${chainId}/tasks`, request)
	return response.data
}

export async function uploadStepAttachment(stepId: number, file: File): Promise<ITaskChainStepAttachment> {
	const formData = new FormData()
	formData.append('files', file)
	const response = await http.put(`/chainsteps/${stepId}/attachments`, formData, {
		headers: {'Content-Type': 'multipart/form-data'},
	})
	return response.data
}

export async function deleteStepAttachment(stepId: number, attachmentId: number): Promise<void> {
	await http.delete(`/chainsteps/${stepId}/attachments/${attachmentId}`)
}
