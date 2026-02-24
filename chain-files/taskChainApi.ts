import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'

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
