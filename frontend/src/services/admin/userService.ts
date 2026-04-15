import {HTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'
import type {IUser} from '@/modelTypes/IUser'

export interface AdminUser extends IUser {
	status: number
	isAdmin: boolean
}

export async function listAdminUsers(params: {s?: string; page?: number; perPage?: number} = {}): Promise<AdminUser[]> {
	const {data} = await HTTPFactory().get('/admin/users', {
		params: objectToSnakeCase(params),
	})
	return (data as unknown[]).map(u => objectToCamelCase(u as Record<string, unknown>)) as AdminUser[]
}

export async function setAdmin(id: number, isAdmin: boolean): Promise<AdminUser> {
	const {data} = await HTTPFactory().patch(`/admin/users/${id}/admin`, {is_admin: isAdmin})
	return objectToCamelCase(data) as unknown as AdminUser
}
