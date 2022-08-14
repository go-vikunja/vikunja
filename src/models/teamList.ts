import TeamShareBaseModel from './teamShareBase'
import type { IList } from './list'

export interface ITeamList extends TeamShareBaseModel {
	listId: IList['id']
}

export default class TeamListModel extends TeamShareBaseModel implements ITeamList {
	listId: IList['id'] = 0

	constructor(data: Partial<ITeamList>) {
		super(data)
		this.assignData(data)
	}
}