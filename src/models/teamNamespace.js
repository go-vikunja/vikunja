import TeamShareBaseModel from './teamShareBase'
import merge from 'lodash/merge'

export default class TeamNamespaceModel extends TeamShareBaseModel {
	defaults() {
		return merge(
			super.defaults(),
			{
				namespaceId: 0,
			},
		)
	}
}