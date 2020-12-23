import {LOADING} from './mutation-types'

export const setLoading = (context, loadFunc = null) => {
	const timeout = setTimeout(() => {
		if (loadFunc === null) {
			context.commit(LOADING, true, {root: true})
		} else {
			loadFunc(true)
		}
	}, 100)
	return () => {
		clearTimeout(timeout)
		if (loadFunc === null) {
			context.commit(LOADING, false, {root: true})
		} else {
			loadFunc(false)
		}
	}
}