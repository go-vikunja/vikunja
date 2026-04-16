import AbstractService from '@/services/abstractService'
import AdminUserModel from '@/models/adminUser'
import type {IAdminUser} from '@/modelTypes/IAdminUser'

export interface CreateAdminUserBody {
	username: string
	email: string
	name?: string
	password?: string
	language?: string
	isAdmin?: boolean
	skipEmailConfirm?: boolean
}

export default class AdminUserService extends AbstractService<IAdminUser> {
	constructor() {
		super({
			getAll: '/admin/users',
			delete: '/admin/users/{id}',
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

	async createUser(body: CreateAdminUserBody) {
		const {data} = await this.http.post('/register', body)
		return this.modelCreateFactory(data)
	}
}
