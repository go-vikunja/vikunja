import Vue from 'vue'

export default {
	namespaced: true,
	// The state is an object which has the list ids as keys.
	state: () => ({}),
	mutations: {
		addList(state, list) {
			Vue.set(state, list.id, list)
		},
		addLists(state, lists) {
			lists.forEach(l => {
				Vue.set(state, l.id, l)
			})
		},
	},
	getters: {
		getListById: state => id => {
			if(typeof state[id] !== 'undefined') {
				return state[id]
			}
			return null
		},
	},
}