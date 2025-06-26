import {ref, computed, readonly} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'
import {findIndexById} from '@/helpers/utils'

import type {IAttachment} from '@/modelTypes/IAttachment'

export const useAttachmentStore = defineStore('attachment', () => {
	const attachments = ref<IAttachment[]>([])

	function set(newAttachments: IAttachment[]) {
		console.debug('Set attachments', newAttachments)
		attachments.value = newAttachments
	}

	function add(attachment: IAttachment) {
		console.debug('Add attachement', attachment)
		attachments.value.push(attachment)
	}

	function removeById(id: IAttachment['id']) {
		const attachmentIndex = findIndexById(attachments.value, id)
		attachments.value.splice(attachmentIndex, 1)
		console.debug('Remove attachement', id)
	}

	const hasAttachments = computed(() => attachments.value.length > 0)

	return {
		attachments: readonly(attachments),
		set,
		add,
		removeById,
		hasAttachments,
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useAttachmentStore, import.meta.hot))
}
