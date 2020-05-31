import AbstractModel from './abstractModel'

export default class BackgroundImageModel extends AbstractModel {
	defaults() {
		return {
			id: 0,
			url: '',
			thumb: '',
			info: {},
		}
	}
}