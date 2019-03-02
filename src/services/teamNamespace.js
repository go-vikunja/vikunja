import AbstractService from './abstractService'
import TeamNamespaceModel from '../models/teamNamespace'
import TeamModel from '../models/team'

export default class TeamNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/teams',
			getAll: '/namespaces/{namespaceID}/teams',
			update: '/namespaces/{namespaceID}/teams/{teamID}',
			delete: '/namespaces/{namespaceID}/teams/{teamID}',
		})
	}

	modelFactory(data) {
		return new TeamNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}
}