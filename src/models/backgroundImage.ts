import AbstractModel, { type IAbstract } from './abstractModel'

export interface IBackgroundImage extends IAbstract {
	id: number
	url: string
	thumb: string
	info: {
		author: string
		authorName: string
	}
	blurHash: string  
}

export default class BackgroundImageModel extends AbstractModel implements IBackgroundImage {
	declare id: number
	declare url: string
	declare thumb: string
	declare info: {
		author: string
		authorName: string
	}
	declare blurHash: string  

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