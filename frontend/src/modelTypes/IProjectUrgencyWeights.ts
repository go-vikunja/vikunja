import type {IAbstract} from './IAbstract'

export interface IProjectUrgencyWeights extends IAbstract {
	urgencyWeights: UrgencyWeight[]
}

export interface IUrgencyWeight {
	property: string
	weight: number
	filter: IBasicFilter
}

export interface IBasicFilter {
	query: string
	includeNulls: boolean
}
