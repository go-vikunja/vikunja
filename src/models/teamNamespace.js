import TeamShareBaseModel from './teamShareBase'
import {merge} from 'lodash'

export default class TeamNamespaceModel extends TeamShareBaseModel {
	defaults() {
		return merge(
			super.defaults(),
			{
				namespaceID: 0,
			}
		)
	}
}