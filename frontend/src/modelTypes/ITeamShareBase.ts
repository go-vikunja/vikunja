import type {IAbstract} from './IAbstract'
import type {ITeam} from './ITeam'
import type {Right} from '@/constants/rights'

export interface ITeamShareBase extends IAbstract {
	teamId: ITeam['id']
	right: Right

	created: Date
	updated: Date
}
