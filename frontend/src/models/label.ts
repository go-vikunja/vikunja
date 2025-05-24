import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'

import {colorIsDark} from '@/helpers/color/colorIsDark'

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

		if (this.hexColor !== '' && !this.hexColor.startsWith('#') && !this.hexColor.startsWith('var(')) {
			this.hexColor = '#' + this.hexColor
		}

		if (this.hexColor === '') {
			this.hexColor = 'var(--grey-200)'
			this.textColor = 'var(--grey-800)'
		} else {
			this.textColor = colorIsDark(this.hexColor)
				// Fixed colors to avoid flipping in dark mode
				? 'hsl(215, 27.9%, 16.9%)' // grey-800
				: 'hsl(220, 13%, 91%)' // grey-200
		}

		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
