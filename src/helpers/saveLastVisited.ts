const LAST_VISITED_KEY = 'lastVisited'

export const saveLastVisited = (name: string, params: object) => {
	localStorage.setItem(LAST_VISITED_KEY, JSON.stringify({name, params}))
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
