import TeamShareBaseModel from './teamShareBase'
import type NamespaceModel from './namespace'

export default class TeamNamespaceModel extends TeamShareBaseModel {
	namespaceId: NamespaceModel['id']

	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}