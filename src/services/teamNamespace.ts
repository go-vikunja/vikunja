import AbstractService from './abstractService'
import TeamNamespaceModel from '@/models/teamNamespace'
import type {ITeamNamespace} from '@/modelTypes/ITeamNamespace'
import TeamModel from '@/models/team'
import {formatISO} from 'date-fns'

export default class TeamNamespaceService extends AbstractService<ITeamNamespace> {
	constructor() {
		super({
			create: '/namespaces/{namespaceId}/teams',
			getAll: '/namespaces/{namespaceId}/teams',
			update: '/namespaces/{namespaceId}/teams/{teamId}',
			delete: '/namespaces/{namespaceId}/teams/{teamId}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new TeamNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}
}