import TeamShareBaseModel from './teamShareBase'

import type {ITeamList} from '@/modelTypes/ITeamList'
import type {IList} from '@/modelTypes/IList'

export default class TeamListModel extends TeamShareBaseModel implements ITeamList {
	listId: IList['id'] = 0

	constructor(data: Partial<ITeamList>) {
		super(data)
		this.assignData(data)
	}
}