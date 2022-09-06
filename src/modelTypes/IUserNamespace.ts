import type {IUserShareBase} from './IUserShareBase'
import type {INamespace} from './INamespace'

export interface IUserNamespace extends IUserShareBase {
	namespaceId: INamespace['id']
}