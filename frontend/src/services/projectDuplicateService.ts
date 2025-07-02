import AbstractService from './abstractService'
import projectDuplicateModel from '@/models/projectDuplicateModel'
import type {IProjectDuplicate} from '@/modelTypes/IProjectDuplicate'

export default class ProjectDuplicateService extends AbstractService<IProjectDuplicate> {
	constructor() {
		super({
			create: '/projects/{projectId}/duplicate',
		})
	}

	beforeCreate(model: IProjectDuplicate): IProjectDuplicate {
		const updatedModel = {
			...model,
			duplicatedProject: null
		}
		return updatedModel
	}

	modelFactory(data: Partial<IProjectDuplicate>): IProjectDuplicate {
		return new projectDuplicateModel(data)
	}
}
