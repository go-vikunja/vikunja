import type {IAbstract} from './IAbstract'

export type AvatarProvider = 'default' | 'initials' | 'gravatar' | 'marble' | 'upload'

export interface IAvatar extends IAbstract {
	avatarProvider: AvatarProvider
}
