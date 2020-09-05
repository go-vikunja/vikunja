import AbstractModel from './abstractModel'

export default class FileModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			mime: '',
			name: '',
			size: '',
			created: null,
		}
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
