const LAST_VISITED_KEY = 'lastVisited'

export const saveLastVisited = (name: string | undefined, params: object, query: object) => {
	if (typeof name === 'undefined') {
		return
	}
	
	localStorage.setItem(LAST_VISITED_KEY, JSON.stringify({name, params, query}))
}

export const getLastVisited = () => {
	const lastVisited = localStorage.getItem(LAST_VISITED_KEY)
	if (lastVisited === null) {
		return null
	}

	return JSON.parse(lastVisited)
}

export const clearLastVisited = () => {
	return localStorage.removeItem(LAST_VISITED_KEY)
}
