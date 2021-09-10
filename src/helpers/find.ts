export function findIndexById(array : [], id : string | number) {
	return array.findIndex(({id: currentId}) => currentId === id)
}