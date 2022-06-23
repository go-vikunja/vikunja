import AbstractModel from './abstractModel'
import TaskModel from './task'
import UserModel from './user'
import type NamespaceModel from './namespace'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'
import SubscriptionModel from '@/models/subscription'

export default class ListModel extends AbstractModel {
	id: number
	title: string
	description: string
	owner: UserModel
	tasks: TaskModel[]
	namespaceId: NamespaceModel['id']
	isArchived: boolean
	hexColor: string
	identifier: string
	backgroundInformation: any
	isFavorite: boolean
	subscription: SubscriptionModel
	position: number
	backgroundBlurHash: string
	
	created: Date
	updated: Date

	constructor(data) {
		super(data)

		this.owner = new UserModel(this.owner)

		// Make all tasks to task models
		this.tasks = this.tasks.map(t => {
			return new TaskModel(t)
		})

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		if (typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			owner: UserModel,
			tasks: [],
			namespaceId: 0,
			isArchived: false,
			hexColor: '',
			identifier: '',
			backgroundInformation: null,
			isFavorite: false,
			subscription: null,
			position: 0,
			backgroundBlurHash: '',

			created: null,
			updated: null,
		}
	}

	isSavedFilter() {
		return this.getSavedFilterId() > 0
	}

	getSavedFilterId() {
		return getSavedFilterIdFromListId(this.id)
	}
}