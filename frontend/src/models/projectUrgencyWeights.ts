import AbstractModel from './abstractModel'

import type {IProjectUrgencyWeights} from '@/modelTypes/IProjectUrgencyWeights'

export default class ProjectUrgencyWeightsModel extends AbstractModel<IProjectUrgencyWeights> implements IProjectUrgencyWeights {
	urgencyWeights: []

	constructor(data: Partial<IProjectUrgencyWeights> = {}) {
		super()
		this.assignData(data)
	}
}
