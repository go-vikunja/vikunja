import AbstractService from './abstractService'
import TeamProjectModel from '@/models/teamProject'
import type {ITeamProject} from '@/modelTypes/ITeamProject'
import type {ITeam} from '@/modelTypes/ITeam'
import TeamModel from '@/models/team'

export default class TeamProjectService extends AbstractService<ITeamProject> {
	constructor() {
		super({
			create: '/projects/{projectId}/teams',
			getAll: '/projects/{projectId}/teams',
			update: '/projects/{projectId}/teams/{teamId}',
			delete: '/projects/{projectId}/teams/{teamId}',
		})
	}

	modelFactory(data: Partial<ITeamProject>): ITeamProject {
		return new TeamProjectModel(data)
	}

	modelGetAllFactory(data: Partial<ITeam>): ITeam {
		return new TeamModel(data)
	}
}
