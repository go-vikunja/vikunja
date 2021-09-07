export function findIndexById(array : [], id : string | number) {
	return array.findIndex(({id: currentId}) => currentId === id)
}

export function findById(array : [], id : string | number) {
	return array.find(({id: currentId}) => currentId === id)
}