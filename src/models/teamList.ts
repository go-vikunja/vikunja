import TeamShareBaseModel from './teamShareBase'
import type { IList } from './list'

export interface ITeamList extends TeamShareBaseModel {
	listId: IList['id']
}

export default class TeamListModel extends TeamShareBaseModel implements ITeamList {
	listId!: IList['id']

	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}