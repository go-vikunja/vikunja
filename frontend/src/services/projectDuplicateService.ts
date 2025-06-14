import AbstractService from './abstractService'
import projectDuplicateModel from '@/models/projectDuplicateModel'
import type {IProjectDuplicate} from '@/modelTypes/IProjectDuplicate'

export default class ProjectDuplicateService extends AbstractService<IProjectDuplicate> {
	constructor() {
		super({
			create: '/projects/{projectId}/duplicate',
		})
	}

	beforeCreate(model) {

		model.project = null
		return model
	}

	modelFactory(data) {
		return new projectDuplicateModel(data)
	}
}
