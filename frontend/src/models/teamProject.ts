import TeamShareBaseModel from './teamShareBase'

import type {ITeamProject} from '@/modelTypes/ITeamProject'
import type {IProject} from '@/modelTypes/IProject'

export default class TeamProjectModel extends TeamShareBaseModel implements ITeamProject {
	projectId: IProject['id'] = 0

	constructor(data: Partial<ITeamProject>) {
		super(data)
		this.assignData(data)
	}
}
