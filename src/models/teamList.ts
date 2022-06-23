import TeamShareBaseModel from './teamShareBase'
import type ListModel from './list'

export default class TeamListModel extends TeamShareBaseModel {
	listId: ListModel['id']

	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}