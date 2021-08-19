import {findIndexById} from '@/helpers/find'

export default {
	namespaced: true,
	state: () => ({
		attachments: [],
	}),
	mutations: {
		set(state, attachments) {
			console.debug('Set attachments', attachments)
			state.attachments = attachments
		},
		add(state, attachment) {
			console.debug('Add attachement', attachment)
			state.attachments.push(attachment)
		},
		removeById(state, id) {
			const attachmentIndex = findIndexById(state.attachments, id)
			state.attachments.splice(attachmentIndex, 1)
			console.debug('Remove attachement', id)
		},
	},
}