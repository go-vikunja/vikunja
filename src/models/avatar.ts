import AbstractModel from './abstractModel'

export type AvatarProvider = 'default' | 'initials' | 'gravatar' | 'marble' | 'upload'

export interface IAvatar extends AbstractModel {
	avatarProvider: AvatarProvider
}

export default class AvatarModel extends AbstractModel implements IAvatar {
	declare avatarProvider: AvatarProvider

	defaults() {
		return {
			avatarProvider: '',
		}
	}
}