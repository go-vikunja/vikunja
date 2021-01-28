import Vue from 'vue'

export default {
	namespaced: true,
	state: () => ({
		attachments: [],
	}),
	mutations: {
		set(state, attachments) {
			console.debug('Set attachments', attachments)
			Vue.set(state, 'attachments', attachments)
		},
		add(state, attachment) {
			console.debug('Add attachement', attachment)
			state.attachments.push(attachment)
		},
		removeById(state, id) {
			for (const a in state.attachments) {
				if (state.attachments[a].id === id) {
					state.attachments.splice(a, 1)
					console.debug('Remove attachement', id)
					break
				}
			}
		},
	},
}