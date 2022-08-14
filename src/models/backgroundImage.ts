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
	id = 0
	url = ''
	thumb = ''
	info: {
		author: string
		authorName: string
	} = {}
	blurHash = ''

	constructor(data: Partial<IBackgroundImage>) {
		super()
		this.assignData(data)
	}
}