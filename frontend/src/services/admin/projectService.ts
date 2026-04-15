import {HTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase} from '@/helpers/case'
import type {IProject} from '@/modelTypes/IProject'

export async function listAdminProjects(params: {page?: number; perPage?: number} = {}): Promise<IProject[]> {
	const {data} = await HTTPFactory().get('/admin/projects', {
		params: {page: params.page, per_page: params.perPage},
	})
	return (data as unknown[]).map(p => objectToCamelCase(p as Record<string, unknown>)) as unknown as IProject[]
}

export async function reassignProjectOwner(projectId: number, newOwnerId: number): Promise<IProject> {
	const {data} = await HTTPFactory().patch(`/admin/projects/${projectId}/owner`, {owner_id: newOwnerId})
	return objectToCamelCase(data) as unknown as IProject
}
