import AbstractModel from './abstractModel'
import type {IFile} from '@/modelTypes/IFile'

export default class FileModel extends AbstractModel<IFile> implements IFile {
	id = 0
	mime = ''
	name = ''
	size = 0
	created: Date = new Date()

	constructor(data: Partial<IFile> = {}) {
		super()
		this.assignData(data)

		this.created = this.created ? new Date(this.created) : new Date()
	}
}
