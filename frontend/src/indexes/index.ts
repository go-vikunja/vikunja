import {Document} from 'flexsearch'

export interface withId {
	id: number,
}

const indexes: { [k: string]: Document<withId> } = {}

export const createNewIndexer = (name: string, fieldsToIndex: string[]) => {
	if (typeof indexes[name] === 'undefined') {
		indexes[name] = new Document<withId>({
			tokenize: 'full',
			document: {
				id: 'id',
				index: fieldsToIndex,
			},
		})
	}

	const index = indexes[name]

	function add(item: withId) {
		return index.add(item.id, item)
	}

	function remove(item: withId) {
		return index.remove(item.id)
	}

	function update(item: withId) {
		return index.update(item.id, item)
	}

	function search(query: string | null) {
		if (query === '' || query === null) {
			return null
		}

		return index.search(query)
			?.flatMap(r => r.result)
			.filter((value, index, self) => self.indexOf(value) === index) as number[]
			|| null
	}

	return {
		add,
		remove,
		update,
		search,
	}
}
