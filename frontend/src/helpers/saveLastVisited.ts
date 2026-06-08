const LAST_VISITED_KEY = 'lastVisited'
const LAST_VISITED_PAGE_KEY = 'lastVisitedPage'

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

export const saveLastVisitedPage = (name: string | undefined, params: object, query: object) => {
	if (typeof name === 'undefined') {
		return
	}

	localStorage.setItem(LAST_VISITED_PAGE_KEY, JSON.stringify({name, params, query}))
}

export const getLastVisitedPage = () => {
	const lastVisited = localStorage.getItem(LAST_VISITED_PAGE_KEY)
	if (lastVisited === null) {
		return null
	}

	return JSON.parse(lastVisited)
}
