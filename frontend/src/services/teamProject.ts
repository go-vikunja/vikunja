import AbstractService from './abstractService'
import TeamProjectModel from '@/models/teamProject'
import type {ITeamProject} from '@/modelTypes/ITeamProject'
import TeamModel from '@/models/team'

export default class TeamProjectService extends AbstractService<ITeamProject> {
	constructor() {
		super({
			create: '/api/v2/projects/{projectId}/teams',
			getAll: '/api/v2/projects/{projectId}/teams',
			update: '/api/v2/projects/{projectId}/teams/{teamId}',
			delete: '/api/v2/projects/{projectId}/teams/{teamId}',
		})
	}

	modelFactory(data) {
		return new TeamProjectModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}

	/**
	 * Performs a post request to the url specified before
	 * @returns {Promise<any | never>}
	 */
	async create(model : ITeamProject) {
		if (this.paths.create === '') {
			throw new Error('This model is not able to create data.')
		}

		const cancel = this.setLoading()
		const finalUrl = this.getReplacedRoute(this.paths.create, model)

		try {
			const response = await this.http.post(finalUrl, model)
			const result = this.modelCreateFactory(response.data)
			if (typeof model.maxPermission !== 'undefined') {
				result.maxPermission = model.maxPermission
			}
			return result
		} finally {
			cancel()
		}
	}
}
