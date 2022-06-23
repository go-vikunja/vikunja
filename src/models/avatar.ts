import AbstractModel from './abstractModel'

export type AVATAR_PROVIDERS = 'default' | 'initials' | 'gravatar' | 'marble' | 'upload'

export default class AvatarModel extends AbstractModel {
	avatarProvider: AVATAR_PROVIDERS

	defaults() {
		return {
			avatarProvider: '',
		}
	}
}