import type {TypeOf} from 'zod'
import {z, string, object} from 'zod'

export const AVATAR_PROVIDER = [
	'default',
	'initials',
	'gravatar',
	'marble',
	'upload',
] as const
export const AvatarProviderSchema = z.enum(AVATAR_PROVIDER)
export type IAvatarProvider = TypeOf<typeof AvatarProviderSchema>

export const AvatarSchema = object({
	// FIXME: shouldn't the default be 'default'?
  avatarProvider: string().or(AvatarProviderSchema).default(''),
})
export type IAvatar = TypeOf<typeof AvatarSchema>