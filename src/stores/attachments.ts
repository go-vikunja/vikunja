import {defineStore, acceptHMRUpdate} from 'pinia'
import {findIndexById} from '@/helpers/utils'

import type {IAttachment} from '@/modelTypes/IAttachment'

export interface AttachmentState {
	attachments: IAttachment[],
}

export const useAttachmentStore = defineStore('attachment', {
	state: (): AttachmentState => ({
		attachments: [],
	}),

	actions: {
		set(attachments: IAttachment[]) {
			console.debug('Set attachments', attachments)
			this.attachments = attachments
		},

		add(attachment: IAttachment) {
			console.debug('Add attachement', attachment)
			this.attachments.push(attachment)
		},

		removeById(id: IAttachment['id']) {
			const attachmentIndex = findIndexById(this.attachments, id)
			this.attachments.splice(attachmentIndex, 1)
			console.debug('Remove attachement', id)
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useAttachmentStore, import.meta.hot))
}