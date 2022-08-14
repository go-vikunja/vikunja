import AbstractModel, { type IAbstract } from '@/models/abstractModel'
import TaskModel, { type ITask } from '@/models/task'
import UserModel, { type IUser } from '@/models/user'
import SubscriptionModel, { type ISubscription } from '@/models/subscription'
import type { INamespace } from '@/models/namespace'

import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'

export interface IList extends IAbstract {
	id: number
	title: string
	description: string
	owner: IUser
	tasks: ITask[]
	namespaceId: INamespace['id']
	isArchived: boolean
	hexColor: string
	identifier: string
	backgroundInformation: any // FIXME: improve type
	isFavorite: boolean
	subscription: ISubscription
	position: number
	backgroundBlurHash: string
	
	created: Date
	updated: Date
}

export default class ListModel extends AbstractModel implements IList {
	id = 0
	title = ''
	description = ''
	owner: IUser = UserModel
	tasks: ITask[] = []
	namespaceId: INamespace['id'] = 0
	isArchived = false
	hexColor = ''
	identifier = ''
	backgroundInformation: any = null
	isFavorite = false
	subscription: ISubscription = null
	position = 0
	backgroundBlurHash = ''
	
	created: Date = null
	updated: Date = null

	constructor(data: Partial<IList>) {
		super()
		this.assignData(data)

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

	isSavedFilter() {
		return this.getSavedFilterId() > 0
	}

	getSavedFilterId() {
		return getSavedFilterIdFromListId(this.id)
	}
}