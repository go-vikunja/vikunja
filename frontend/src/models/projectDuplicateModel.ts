import AbstractModel from './abstractModel'
import ProjectModel from './project'

import type {IProjectDuplicate} from '@/modelTypes/IProjectDuplicate'
import type {IProject} from '@/modelTypes/IProject'

export default class ProjectDuplicateModel extends AbstractModel<IProjectDuplicate> implements IProjectDuplicate {
	projectId = 0
	duplicatedProject: IProject | null = null
	parentProjectId = 0

	constructor(data : Partial<IProjectDuplicate>) {
		super()
		this.assignData(data)

		this.duplicatedProject = this.duplicatedProject ? new ProjectModel(this.duplicatedProject) : null
	}
}
