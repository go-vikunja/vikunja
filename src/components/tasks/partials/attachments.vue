<template>
	<div class="attachments">
		<h3>
			<span class="icon is-grey">
				<icon icon="paperclip"/>
			</span>
			Attachments
			<a
					class="button is-primary is-outlined is-small noshadow"
					@click="$refs.files.click()"
					:disabled="attachmentService.loading">
				<span class="icon is-small"><icon icon="cloud-upload-alt"/></span>
				Upload attachment
			</a>
		</h3>

		<input type="file" id="files" ref="files" multiple @change="uploadNewAttachment()" :disabled="attachmentService.loading"/>
		<progress v-if="attachmentService.uploadProgress > 0" class="progress is-primary" :value="attachmentService.uploadProgress" max="100">{{ attachmentService.uploadProgress }}%</progress>

		<table>
			<tr>
				<th>Name</th>
				<th>Size</th>
				<th>Type</th>
				<th>Date</th>
				<th>Created By</th>
				<th>Action</th>
			</tr>
			<tr class="attachment" v-for="a in attachments" :key="a.id">
				<td>
					{{ a.file.name }}
				</td>
				<td>{{ a.file.getHumanSize() }}</td>
				<td>{{ a.file.mime }}</td>
				<td v-tooltip="formatDate(a.created)">{{ formatDateSince(a.created) }}</td>
				<td><user :user="a.createdBy" :avatar-size="30"/></td>
				<td>
					<div class="buttons has-addons">
						<a class="button is-primary noshadow" @click="downloadAttachment(a)" v-tooltip="'Download this attachment'">
							<span class="icon">
								<icon icon="cloud-download-alt"/>
							</span>
						</a>
						<a class="button is-danger noshadow" v-tooltip="'Delete this attachment'" @click="() => {attachmentToDelete = a; showDeleteModal = true}">
							<span class="icon">
								<icon icon="trash-alt"/>
							</span>
						</a>
					</div>
				</td>
			</tr>
		</table>

		<!-- Dropzone -->
		<div class="dropzone" :class="{ 'hidden': !showDropzone }">
			<div class="drop-hint">
				<div class="icon">
					<icon icon="cloud-upload-alt"/>
				</div>
				<div class="hint">
					Drop files here to upload
				</div>
			</div>
		</div>

		<!-- Delete modal -->
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteAttachment()">
			<span slot="header">Delete attachment</span>
			<p slot="text">Are you sure you want to delete the attachment {{ attachmentToDelete.file.name }}?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import AttachmentService from '../../../services/attachment'
	import AttachmentModel from '../../../models/attachment'
	import User from '../../misc/user'

	export default {
		name: 'attachments',
		components: {
			User,
		},
		data() {
			return {
				attachments: [],
				attachmentService: AttachmentService,
				showDropzone: false,

				showDeleteModal: false,
				attachmentToDelete: AttachmentModel,
			}
		},
		props: {
			taskId: {
				required: true,
				type: Number,
			},
			initialAttachments: {
				type: Array,
			}
		},
		created() {
			this.attachmentService = new AttachmentService()
			this.attachments = this.initialAttachments
		},
		mounted() {
			document.addEventListener('dragenter', e => {
				e.stopPropagation()
				e.preventDefault()
				this.showDropzone = true
			});

			window.addEventListener('dragleave', e => {
				e.stopPropagation()
				e.preventDefault()
				this.showDropzone = false
			});

			document.addEventListener('dragover', e => {
				e.stopPropagation()
				e.preventDefault()
				this.showDropzone = true
			});

			document.addEventListener('drop', e => {
				e.stopPropagation()
				e.preventDefault()

				let files = e.dataTransfer.files
				this.uploadFiles(files)
				this.showDropzone = false
			})
		},
		watch: {
			initialAttachments(newVal) {
				this.attachments = newVal
			},
		},
		methods: {
			downloadAttachment(attachment) {
				this.attachmentService.download(attachment)
			},
			uploadNewAttachment() {
				if(this.$refs.files.files.length === 0) {
					return
				}

				this.uploadFiles(this.$refs.files.files)
			},
			uploadFiles(files) {
				const attachmentModel = new AttachmentModel({taskId: this.taskId})
				this.attachmentService.create(attachmentModel, files)
					.then(r => {
						if(r.success !== null) {
							r.success.forEach(a => {
								this.attachments.push(a)
								this.$store.dispatch('tasks/addTaskAttachment', {taskId: this.taskId, attachment: a})
							})
						}
						if(r.errors !== null) {
							r.errors.forEach(m => {
								this.error(m)
							})
						}
					})
					.catch(e => {
						this.error(e, this)
					})

			},
			deleteAttachment() {
				this.attachmentService.delete(this.attachmentToDelete)
					.then(r => {
						// Remove the file from the list
						for (const a in this.attachments) {
							if (this.attachments[a].id === this.attachmentToDelete.id) {
								this.attachments.splice(a, 1)
							}
						}
						this.success(r, this)
					})
					.catch(e => {
						this.error(e, this)
					})
					.finally(() => {
						this.showDeleteModal = false
					})
			},
		},
	}
</script>
