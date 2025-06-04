import type {IAbstract} from './IAbstract'

export type AvatarProvider = 'default' | 'initials' | 'gravatar' | 'marble' | 'upload' | 'ldap'

export interface IAvatar extends IAbstract {
	avatarProvider: AvatarProvider
}
