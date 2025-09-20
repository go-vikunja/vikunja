import AbstractService from './abstractService'
import projectDuplicateModel from '@/models/projectDuplicateModel'
import type {IProjectDuplicate} from '@/modelTypes/IProjectDuplicate'

export default class ProjectDuplicateService extends AbstractService<IProjectDuplicate> {
	constructor() {
		super({
			create: '/projects/{projectId}/duplicate',
		})
	}

	beforeCreate(model: IProjectDuplicate) {

		model.project = null
		return model
	}

	modelFactory(data: Partial<IProjectDuplicate>) {
		return new projectDuplicateModel(data)
	}
}
