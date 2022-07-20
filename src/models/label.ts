import AbstractModel from './abstractModel'
import UserModel, { type IUser } from './user'
import {colorIsDark} from '@/helpers/color/colorIsDark'

const DEFAULT_LABEL_BACKGROUND_COLOR = 'e8e8e8'

export interface ILabel {
	id: number
	title: string
	hexColor: string
	description: string
	createdBy: IUser
	listId: number
	textColor: string

	created: Date
	updated: Date
}

export default class LabelModel extends AbstractModel implements ILabel {
	declare id: number
	declare title: string
	declare hexColor: string
	declare description: string
	declare createdBy: IUser
	declare listId: number
	declare textColor: string

	created: Date
	updated: Date

	constructor(data) {
		super(data)
		// FIXME: this should be empty and be definied in the client.
		// that way it get's never send to the server db and is easier to change in future versions.
		// Set the default color
		if (this.hexColor === '') {
			this.hexColor = DEFAULT_LABEL_BACKGROUND_COLOR
		}
		if (this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}
		this.textColor = colorIsDark(this.hexColor) ? '#4a4a4a' : '#fff'
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