import TeamShareBaseModel from './teamShareBase'
import {merge} from 'lodash'

export default class TeamListModel extends TeamShareBaseModel {
	defaults() {
		return merge(
			super.defaults(),
			{
				listId: 0,
			}
		)
	}
}