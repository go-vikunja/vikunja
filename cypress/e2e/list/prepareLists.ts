import {ListFactory} from '../../factories/list'
import {NamespaceFactory} from '../../factories/namespace'
import {TaskFactory} from '../../factories/task'

export function createLists() {
	NamespaceFactory.create(1)
	const lists = ListFactory.create(1, {
		title: 'First List'
	})
	TaskFactory.truncate()
	return lists
}

export function prepareLists(setLists = (...args: any[]) => {}) {
	beforeEach(() => {
		const lists = createLists()
		setLists(lists)
	})
}