import AbstractModel from './abstractModel'
import ProjectModel from './project'

import type {IProjectDuplicate} from '@/modelTypes/IProjectDuplicate'
import type {IProject} from '@/modelTypes/IProject'

export default class ProjectDuplicateModel extends AbstractModel<IProjectDuplicate> implements IProjectDuplicate {
	projectId = 0
	project: IProject = ProjectModel

	constructor(data : Partial<IProjectDuplicate>) {
		super()
		this.assignData(data)

		this.project = new ProjectModel(this.project)
	}
}