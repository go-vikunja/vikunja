import type {IProjectUrgencyWeights, IProjectUrgencyWeight} from '@/modelTypes/IProjectUrgencyWeights'
import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'

export interface ProjectUrgencyWeightListResult {
	items: IProjectUrgencyWeight[]
	total: number
	page: number
	perPage: number
	totalPages: number
}

export function useProjectUrgencyWeightsService() {
	const http = AuthenticatedHTTPFactory()

	async function getAll(params: IProjectUrgencyWeights): Promise<ProjectUrgencyWeightListResult> {
		const {data} = await http.get(apiV2Url(`projects/${encodeURIComponent(params.projectID)}/urgency_weights`), {})
		return objectToCamelCase(data)
	}

	async function updateAll(params: IProjectUrgencyWeights): Promise<IProjectUrgencyWeights> {
		const body = objectToSnakeCase(params)
		delete body.project_id
		const {data} = await http.put(apiV2Url(`projects/${encodeURIComponent(params.projectID)}/urgency_weights`), body)
		return objectToCamelCase(data)
	}

	return {getAll, updateAll}
}
