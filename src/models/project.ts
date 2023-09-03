import AbstractModel from './abstractModel'
import TaskModel from '@/models/task'
import UserModel from '@/models/user'
import SubscriptionModel from '@/models/subscription'

import type {IProject} from '@/modelTypes/IProject'
import type {IUser} from '@/modelTypes/IUser'
import type {ITask} from '@/modelTypes/ITask'
import type {ISubscription} from '@/modelTypes/ISubscription'

export default class ProjectModel extends AbstractModel<IProject> implements IProject {
	id = 0
	title = ''
	description = ''
	owner: IUser = UserModel
	tasks: ITask[] = []
	isArchived = false
	hexColor = ''
	identifier = ''
	backgroundInformation: unknown | null = null
	isFavorite = false
	subscription: ISubscription = null
	position = 0
	backgroundBlurHash = ''
	parentProjectId = 0
	doneBucketId = 0
	defaultBucketId = 0
	
	created: Date = null
	updated: Date = null

	constructor(data: Partial<IProject> = {}) {
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
}