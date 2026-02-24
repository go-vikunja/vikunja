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

export interface ITaskChainStep {
	id?: number
	chain_id?: number
	sequence: number
	title: string
	description: string
	offset_days: number
	duration_days: number
	priority: number
	hex_color: string
	label_ids: number[]
	attachments?: ITaskChainStepAttachment[]
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
	step_description_overrides?: Record<number, string>
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
