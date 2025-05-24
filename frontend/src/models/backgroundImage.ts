import AbstractModel from './abstractModel'
import type {IBackgroundImage} from '@/modelTypes/IBackgroundImage'

export default class BackgroundImageModel extends AbstractModel<IBackgroundImage> implements IBackgroundImage {
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
