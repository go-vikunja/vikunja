import type {IAbstract} from './IAbstract'

export interface IProjectUrgencyWeights extends IAbstract {
	projectID: number
	urgencyWeights: IProjectUrgencyWeight[]
}

export interface IProjectUrgencyWeight {
	property: string
	weight: number
	filter: IBasicFilter
}

export interface IBasicFilter {
	query: string
	includeNulls: boolean
}
