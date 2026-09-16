import AbstractService from '@/services/abstractService'
import AdminUserModel from '@/models/adminUser'
import type {IAdminUser} from '@/modelTypes/IAdminUser'
import {apiV2Url} from '@/helpers/fetcher'

export interface CreateAdminUserBody {
	username: string
	email: string
	password: string
	name?: string
	language?: string
	isAdmin?: boolean
	skipEmailConfirm?: boolean
}

export type DeleteUserMode = 'now' | 'scheduled'

export default class AdminUserService extends AbstractService<IAdminUser> {
	constructor() {
		super({
			getAll: '/admin/users',
		})
	}

	modelFactory(data: Partial<IAdminUser>) {
		return new AdminUserModel(data)
	}

	async setAdmin(id: IAdminUser['id'], isAdmin: boolean) {
		const {data} = await this.http.patch(`/admin/users/${id}/admin`, {is_admin: isAdmin})
		return this.modelUpdateFactory(data)
	}

	async setStatus(id: IAdminUser['id'], status: number) {
		const {data} = await this.http.patch(`/admin/users/${id}/status`, {status})
		return this.modelUpdateFactory(data)
	}

	// The password endpoints only exist on /api/v2, hence the absolute URLs.
	async setPassword(id: IAdminUser['id'], newPassword: string) {
		const {data} = await this.http.patch(apiV2Url(`admin/users/${id}/password`), {new_password: newPassword})
		return this.modelUpdateFactory(data)
	}

	async sendPasswordResetEmail(id: IAdminUser['id']) {
		await this.http.post(apiV2Url(`admin/users/${id}/password-reset-email`))
	}

	async createUser(body: CreateAdminUserBody) {
		const {data} = await this.http.post('/admin/users', body)
		return this.modelCreateFactory(data)
	}

	async deleteUser(id: IAdminUser['id'], mode: DeleteUserMode) {
		await this.http.delete(`/admin/users/${id}`, {params: {mode}})
	}
}
