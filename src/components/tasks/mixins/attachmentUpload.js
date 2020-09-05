import AttachmentModel from '../../../models/attachment'
import AttachmentService from '../../../services/attachment'

export default {
	methods: {
		attachmentUpload(file, onSuccess) {
			const files = [file]

			const attachmentService = new AttachmentService()

			const attachmentModel = new AttachmentModel({taskId: this.taskId})
			attachmentService.create(attachmentModel, files)
				.then(r => {
					if (r.success !== null) {
						r.success.forEach(a => {
							this.$store.commit('attachments/add', a)
							this.$store.dispatch('tasks/addTaskAttachment', {taskId: this.taskId, attachment: a})
							onSuccess(`${window.API_URL}/tasks/${this.taskId}/attachments/${a.id}`)
						})
					}
					if (r.errors !== null) {
						r.errors.forEach(m => {
							this.error(m)
						})
					}
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}