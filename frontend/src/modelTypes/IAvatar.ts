import type {IAbstract} from './IAbstract'

export const AVATAR_PROVIDER = [
	'default',
	'initials',
	'gravatar',
	'marble',
	'upload',
] as const
export type AvatarProvider = typeof AVATAR_PROVIDER[number]

export interface IAvatar extends IAbstract {
	avatarProvider: AvatarProvider
}