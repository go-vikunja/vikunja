import type {IAbstract} from './IAbstract'

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
