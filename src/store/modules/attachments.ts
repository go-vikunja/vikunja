import type { Module } from 'vuex'
import {findIndexById} from '@/helpers/utils'

import type { AttachmentState, RootStoreState } from '@/store/types'
import type { IAttachment } from '@/modelTypes/IAttachment'

const store : Module<AttachmentState, RootStoreState> = {
	namespaced: true,
	state: () => ({
		attachments: [],
	}),
	mutations: {
		set(state, attachments: IAttachment[]) {
			console.debug('Set attachments', attachments)
			state.attachments = attachments
		},
		add(state, attachment: IAttachment) {
			console.debug('Add attachement', attachment)
			state.attachments.push(attachment)
		},
		removeById(state, id: IAttachment['id']) {
			const attachmentIndex = findIndexById<IAttachment>(state.attachments, id)
			state.attachments.splice(attachmentIndex, 1)
			console.debug('Remove attachement', id)
		},
	},
}

export default store