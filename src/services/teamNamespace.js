import AbstractService from './abstractService'
import TeamNamespaceModel from '../models/teamNamespace'
import TeamModel from '../models/team'
import {formatISO} from 'date-fns'

export default class TeamNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceId}/teams',
			getAll: '/namespaces/{namespaceId}/teams',
			update: '/namespaces/{namespaceId}/teams/{teamId}',
			delete: '/namespaces/{namespaceId}/teams/{teamId}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new TeamNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}
}