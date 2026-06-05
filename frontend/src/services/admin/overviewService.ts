import AbstractService from '@/services/abstractService'
import AdminOverviewModel from '@/models/adminOverview'
import type {IAdminOverview} from '@/modelTypes/IAdminOverview'

export default class AdminOverviewService extends AbstractService<IAdminOverview> {
	modelFactory(data: Partial<IAdminOverview>) {
		return new AdminOverviewModel(data)
	}

	async getOverview() {
		const {data} = await this.http.get('/admin/overview')
		return this.modelGetFactory(data)
	}
}
