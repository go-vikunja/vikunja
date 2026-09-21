import type {User} from '@/client/generated'
import {avatarKeys} from '@/client/queries/avatars'
import {queryClient} from '@/client/queryClient'

export function invalidateAvatarCache(user: Pick<User, 'username'>) {
	if (user?.username) void queryClient.invalidateQueries({queryKey: avatarKeys.user(user.username)})
}

export type UserWithId = User & {id: number}

export function getDisplayName(user: Pick<User, 'name' | 'username'> | null | undefined) {
	return user?.name || user?.username || ''
}

