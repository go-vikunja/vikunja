import type {ITeamShareBase} from './ITeamShareBase'
import type {INamespace} from './INamespace'

export interface ITeamNamespace extends ITeamShareBase {
	namespaceId: INamespace['id']
}