const DEFAULT_ID_LENGTH = 9

export function createRandomID(idLength = DEFAULT_ID_LENGTH) {
	return Math.random().toString(36).slice(2, idLength)
}
