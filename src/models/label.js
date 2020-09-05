import AbstractModel from './abstractModel'
import UserModel from './user'
import {colorIsDark} from '@/helpers/colorIsDark'

export default class LabelModel extends AbstractModel {
	constructor(data) {
		super(data)
		// Set the default color
		if (this.hexColor === '') {
			this.hexColor = 'e8e8e8'
		}
		if (this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}
		this.textColor = colorIsDark(this.hexColor) ? '#4a4a4a' : '#e5e5e5'
		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			title: '',
			hexColor: '',
			description: '',
			createdBy: UserModel,
			listId: 0,
			textColor: '',

			created: null,
			updated: null,
		}
	}
}