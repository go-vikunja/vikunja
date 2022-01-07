import {LOADING, LOADING_MODULE} from './mutation-types'

/**
 * This helper sets the loading state with a 100ms delay to avoid flickering.
 * 
 * @param {*} context The vuex module context.
 * @param {null|String} module The module that is loading. This parameter allows components to listen for specific parts of the application loading.
 * @param {null|function} loadFunc If not null, this function will be executed instead of the default setting loading.
 */
export const setLoading = (context, module = null, loadFunc = null) => {
	const timeout = setTimeout(() => {
		if (loadFunc === null) {
			context.commit(LOADING, true, {root: true})
			context.commit(LOADING_MODULE, module, {root: true})
		} else {
			loadFunc(true)
		}
	}, 100)
	return () => {
		clearTimeout(timeout)
		if (loadFunc === null) {
			context.commit(LOADING, false, {root: true})
			context.commit(LOADING_MODULE, null, {root: true})
		} else {
			loadFunc(false)
		}
	}
}