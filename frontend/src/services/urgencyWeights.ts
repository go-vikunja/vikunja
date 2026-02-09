import AbstractService from './abstractService'
import ProjectUrgencyWeightsModel from '@/models/projectUrgencyWeights'
import type {IProjectUrgencyWeights} from '@/modelTypes/IProjectUrgencyWeights'

export default class ProjectUrgencyWeightsService extends AbstractService<IProjectUrgencyWeights> {
	constructor() {
		super({
			get: '/projects/{id}/urgency_weights',
			create: '/projects/{id}/urgency_weights',
		})
	}

	modelFactory(data: Partial<IProjectUrgencyWeights>) {
		return new ProjectUrgencyWeightsModel(data)
	}
}
