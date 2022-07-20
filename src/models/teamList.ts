import TeamShareBaseModel from './teamShareBase'
import type { IList } from './list'

export interface ITeamList {
	listId: IList['id']
}

export default class TeamListModel extends TeamShareBaseModel implements ITeamList {
	declare listId: IList['id']

	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}