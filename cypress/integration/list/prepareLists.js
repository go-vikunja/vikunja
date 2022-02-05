import {ListFactory} from '../../factories/list'
import {UserFactory} from '../../factories/user'
import {NamespaceFactory} from '../../factories/namespace'
import {TaskFactory} from '../../factories/task'

export function prepareLists(setLists = () => {}) {
	beforeEach(() => {
		UserFactory.create(1)
		NamespaceFactory.create(1)
		const lists = ListFactory.create(1, {
			title: 'First List'
		})
		setLists(lists)
		TaskFactory.truncate()
	})
}