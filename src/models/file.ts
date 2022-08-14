import AbstractModel, { type IAbstract } from './abstractModel'

export interface IFile extends IAbstract {
	id: number
	mime: string
	name: string
	size: number
	created: Date
} 

export default class FileModel extends AbstractModel implements IFile {
	id = 0
	mime = ''
	name = ''
	size = 0
	created: Date = null

	constructor(data: Partial<IFile>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
	}

	getHumanSize() {
		const sizes = {
			0: 'B',
			1: 'KB',
			2: 'MB',
			3: 'GB',
			4: 'TB',
		}

		let it = 0
		let size = this.size
		while (size > 1024) {
			size /= 1024
			it++
		}

		return Number(Math.round(size + 'e2') + 'e-2') + ' ' + sizes[it]
	}
}
