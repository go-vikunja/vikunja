import TeamShareBaseModel from './teamShareBase'

export default class TeamListModel extends TeamShareBaseModel {
	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}