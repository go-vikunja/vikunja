import type {ITeamShareBase} from './ITeamShareBase'
import type {IList} from './IList'

export interface ITeamList extends ITeamShareBase {
	listId: IList['id']
}