import AbstractModel from './abstractModel'

export default class AvatarModel extends AbstractModel {
	defaults() {
		return {
			avatarProvider: '',
		}
	}
}