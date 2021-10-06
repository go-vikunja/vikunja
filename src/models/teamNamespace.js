import TeamShareBaseModel from './teamShareBase'

export default class TeamNamespaceModel extends TeamShareBaseModel {
	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}