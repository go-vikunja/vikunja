import {LOADING} from './mutation-types'

export const setLoading = context => {
	const timeout = setTimeout(() => {
		context.commit(LOADING, true, {root: true})
	}, 100)
	return () => {
		clearTimeout(timeout)
		context.commit(LOADING, false, {root: true})
	}
}