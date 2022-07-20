import TeamShareBaseModel from './teamShareBase'
import type { INamespace } from './namespace'

export interface ITeamNamespace {
	namespaceId: INamespace['id']
}

export default class TeamNamespaceModel extends TeamShareBaseModel implements ITeamNamespace {
	declare namespaceId: INamespace['id']

	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}