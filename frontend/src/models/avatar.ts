import AbstractModel from './abstractModel'
import type { IAvatar } from '@/modelTypes/IAvatar'

export default class AvatarModel extends AbstractModel<IAvatar> implements IAvatar {
	avatarProvider: IAvatar['avatarProvider'] = 'default'

	constructor(data: Partial<IAvatar>) {
		super()
		this.assignData(data)
	}
}
