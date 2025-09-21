import type {Permission} from '@/constants/permissions'

export interface IAbstract {
	[key: string]: unknown
	maxPermission: Permission | null // FIXME: should this be readonly?
}
