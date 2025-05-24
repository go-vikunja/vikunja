import type {Right} from '@/constants/rights'

export interface IAbstract {
	maxRight: Right | null // FIXME: should this be readonly?
}
