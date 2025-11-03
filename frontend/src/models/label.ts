import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'

import {getTextColor} from '@/helpers/color/getTextColor'

export default class LabelModel extends AbstractModel<ILabel> implements ILabel {
	id = 0
	title = ''
	// FIXME: this should be empty and be defined in the client.
	// that way it gets never send to the server db and is easier to change in future versions.
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

		if (this.hexColor !== '' && !this.hexColor.startsWith('#') && !this.hexColor.startsWith('var(')) {
			this.hexColor = '#' + this.hexColor
		}

		if (this.hexColor !== '') {
			this.textColor = getTextColor(this.hexColor)
		}

		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
