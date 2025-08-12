import type {Permission} from '@/constants/permissions'

export interface IAbstract {
	maxPermission: Permission | null // FIXME: should this be readonly?
}
