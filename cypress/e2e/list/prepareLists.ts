import {ListFactory} from '../../factories/list'
import {UserFactory} from '../../factories/user'
import {NamespaceFactory} from '../../factories/namespace'
import {TaskFactory} from '../../factories/task'

export function createLists() {
	UserFactory.create(1)
	NamespaceFactory.create(1)
	const lists = ListFactory.create(1, {
		title: 'First List'
	})
	TaskFactory.truncate()
	return lists
}

export function prepareLists(setLists = () => {}) {
	beforeEach(() => {
		const lists = createLists()
		setLists(lists)
	})
}