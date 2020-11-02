import TeamShareBaseModel from './teamShareBase'
import merge from 'lodash/merge'

export default class TeamListModel extends TeamShareBaseModel {
	defaults() {
		return merge(
			super.defaults(),
			{
				listId: 0,
			},
		)
	}
}