import AbstractModel from './abstractModel'
import UserModel from '@/models/user'

import type {IWikiPage} from '@/modelTypes/IWikiPage'
import type {IUser} from '@/modelTypes/IUser'

export default class WikiPageModel extends AbstractModel<IWikiPage> implements IWikiPage {
	id = 0
	projectId = 0
	parentId: number | null = null
	title = ''
	content = ''
	path = ''
	isFolder = false
	position = 0
	createdBy: IUser = UserModel
	
	created: Date = null
	updated: Date = null
	
	children?: IWikiPage[] = []

	constructor(data: Partial<IWikiPage> = {}) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)
		
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
		
		if (this.children) {
			this.children = this.children.map(c => new WikiPageModel(c))
		}
	}
}
