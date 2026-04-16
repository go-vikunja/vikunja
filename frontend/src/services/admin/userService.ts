import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'
import type {IUser} from '@/modelTypes/IUser'

export interface AdminUser extends IUser {
	status: number
	isAdmin: boolean
	issuer: string
	subject?: string
	authProvider?: string
}

export async function listAdminUsers(params: {s?: string; page?: number; perPage?: number} = {}): Promise<AdminUser[]> {
	const {data} = await AuthenticatedHTTPFactory().get('/admin/users', {
		params: objectToSnakeCase(params),
	})
	return (data as unknown[]).map(u => objectToCamelCase(u as Record<string, unknown>)) as AdminUser[]
}

export async function setAdmin(id: number, isAdmin: boolean): Promise<AdminUser> {
	const {data} = await AuthenticatedHTTPFactory().patch(`/admin/users/${id}/admin`, {is_admin: isAdmin})
	return objectToCamelCase(data) as unknown as AdminUser
}

export async function setStatus(id: number, status: number): Promise<AdminUser> {
	const {data} = await AuthenticatedHTTPFactory().patch(`/admin/users/${id}/status`, {status})
	return objectToCamelCase(data) as unknown as AdminUser
}

export async function deleteUser(id: number): Promise<void> {
	await AuthenticatedHTTPFactory().delete(`/admin/users/${id}`)
}

export interface CreateAdminUserBody {
	username: string
	email: string
	name?: string
	password?: string
	language?: string
	isAdmin?: boolean
	skipEmailConfirm?: boolean
}

export async function createAdminUser(body: CreateAdminUserBody): Promise<AdminUser> {
	const {data} = await AuthenticatedHTTPFactory().post('/register', objectToSnakeCase(body))
	return objectToCamelCase(data) as unknown as AdminUser
}
