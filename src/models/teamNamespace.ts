import TeamShareBaseModel from './teamShareBase'
import type { INamespace } from './namespace'

export interface ITeamNamespace extends TeamShareBaseModel {
	namespaceId: INamespace['id']
}

export default class TeamNamespaceModel extends TeamShareBaseModel implements ITeamNamespace {
	namespaceId!: INamespace['id']

	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}