import type {User} from '@/client/generated'

export type UserWithId = User & {id: number}

export function getDisplayName(user: Pick<User, 'name' | 'username'> | null | undefined) {
	return user?.name || user?.username || ''
}

