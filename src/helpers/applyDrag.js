export const applyDrag = (arr, dragResult) => {
	const {removedIndex, addedIndex, payload} = dragResult
	if (removedIndex === null && addedIndex === null) return arr

	const result = [...arr]
	// The payload comes from the task itself
	let itemToAdd = payload

	if (removedIndex !== null) {
		itemToAdd = result.splice(removedIndex, 1)[0]
	}

	if (addedIndex !== null) {
		result.splice(addedIndex, 0, itemToAdd)
	}

	return result
}
