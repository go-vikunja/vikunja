import AbstractModel from './abstractModel'
import ProjectModel from './project'

import type {IProjectTemplate} from '@/modelTypes/IProjectTemplate'
import type {IProject} from '@/modelTypes/IProject'

export default class ProjectTemplateModel extends AbstractModel<IProjectTemplate> implements IProjectTemplate {
	projectId = 0
	project: IProject | null = null

	constructor(data: Partial<IProjectTemplate>) {
		super()
		this.assignData(data)

		this.project = this.project ? new ProjectModel(this.project) : null
	}
}
