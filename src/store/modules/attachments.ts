import {findIndexById} from '@/helpers/utils'

import type { AttachmentState } from '@/store/types'
import type { IAttachment } from '@/models/attachment'

export default {
	namespaced: true,
	state: (): AttachmentState => ({
		attachments: [],
	}),
	mutations: {
		set(state: AttachmentState, attachments: IAttachment[]) {
			console.debug('Set attachments', attachments)
			state.attachments = attachments
		},
		add(state: AttachmentState, attachment: IAttachment) {
			console.debug('Add attachement', attachment)
			state.attachments.push(attachment)
		},
		removeById(state: AttachmentState, id: IAttachment['id']) {
			const attachmentIndex = findIndexById<IAttachment>(state.attachments, id)
			state.attachments.splice(attachmentIndex, 1)
			console.debug('Remove attachement', id)
		},
	},
}