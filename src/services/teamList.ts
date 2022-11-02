import AbstractService from './abstractService'
import TeamListModel from '@/models/teamList'
import type {ITeamList} from '@/modelTypes/ITeamList'
import TeamModel from '@/models/team'

export default class TeamListService extends AbstractService<ITeamList> {
	constructor() {
		super({
			create: '/lists/{listId}/teams',
			getAll: '/lists/{listId}/teams',
			update: '/lists/{listId}/teams/{teamId}',
			delete: '/lists/{listId}/teams/{teamId}',
		})
	}

	modelFactory(data) {
		return new TeamListModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}
}