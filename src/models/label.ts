import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'

import {colorIsDark} from '@/helpers/color/colorIsDark'
import {getRandomColorHex} from '@/helpers/color/randomColor'

export default class LabelModel extends AbstractModel<ILabel> implements ILabel {
	id = 0
	title = ''
	// FIXME: this should be empty and be definied in the client.
	// that way it get's never send to the server db and is easier to change in future versions.
	hexColor = ''
	description = ''
	createdBy: IUser
	projectId = 0
	textColor = ''

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ILabel> = {}) {
		super()
		this.assignData(data)
		
		if (this.hexColor === '') {
			this.hexColor = getRandomColorHex()
		}

		if (this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}
		this.textColor = colorIsDark(this.hexColor) ? '#4a4a4a' : '#fff'
		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}