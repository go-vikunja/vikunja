import AbstractService from './abstractService'
import ProjectTemplateModel from '@/models/projectTemplateModel'
import type {IProjectTemplate} from '@/modelTypes/IProjectTemplate'

export default class ProjectTemplateService extends AbstractService<IProjectTemplate> {
	constructor() {
		super({
			create: '/projects/{projectId}/template',
		})
	}

	modelFactory(data: Partial<IProjectTemplate>) {
		return new ProjectTemplateModel(data)
	}
}
