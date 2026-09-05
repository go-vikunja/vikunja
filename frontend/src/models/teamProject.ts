import TeamShareBaseModel from './teamShareBase'

import type {ITeamProject} from '@/modelTypes/ITeamProject'

export default class TeamProjectModel extends TeamShareBaseModel implements ITeamProject {
	projectId = 0

	constructor(data: Partial<ITeamProject>) {
		super(data)
		this.assignData(data)
	}
}
