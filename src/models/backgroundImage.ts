import AbstractModel from './abstractModel'

export default class BackgroundImageModel extends AbstractModel {
	id: number
	url: string
	thumb: string
	info: {
		author: string
		authorName: string
	}
	blurHash: string  

	defaults() {
		return {
			id: 0,
			url: '',
			thumb: '',
			info: {},
			blurHash: '',
		}
	}
}