import type {IAbstract} from './IAbstract'
import type {Team as ITeam} from '@/client/generated'
import type {Permission} from '@/constants/permissions'

export interface ITeamShareBase extends IAbstract {
	teamId: ITeam['id']
	permission: Permission

	created: Date
	updated: Date
}
