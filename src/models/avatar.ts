import AbstractModel, { type IAbstract } from './abstractModel'

export type AvatarProvider = 'default' | 'initials' | 'gravatar' | 'marble' | 'upload'

export interface IAvatar extends IAbstract {
	avatarProvider: AvatarProvider
}

export default class AvatarModel extends AbstractModel implements IAvatar {
	avatarProvider: AvatarProvider = 'default'

	constructor(data: Partial<IAvatar>) {
		super()
		this.assignData(data)
	}
}