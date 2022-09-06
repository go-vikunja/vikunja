import TeamShareBaseModel from './teamShareBase'

import type {ITeamNamespace} from '@/modelTypes/ITeamNamespace'
import type {INamespace} from '@/modelTypes/INamespace'

export default class TeamNamespaceModel extends TeamShareBaseModel implements ITeamNamespace {
	namespaceId: INamespace['id'] = 0

	constructor(data: Partial<ITeamNamespace>) {
		super(data)
		this.assignData(data)
	}
}