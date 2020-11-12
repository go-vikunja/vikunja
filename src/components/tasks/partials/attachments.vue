<template>
	<div class="attachments">
		<h3>
			<span class="icon is-grey">
				<icon icon="paperclip"/>
			</span>
			Attachments
			<a
				:disabled="attachmentService.loading"
				@click="$refs.files.click()"
				class="button is-primary is-outlined is-small noshadow"
				v-if="editEnabled">
				<span class="icon is-small"><icon icon="cloud-upload-alt"/></span>
				Upload attachment
			</a>
		</h3>

		<input
			:disabled="attachmentService.loading"
			@change="uploadNewAttachment()"
			id="files"
			multiple
			ref="files"
			type="file"
			v-if="editEnabled"/>
		<progress
			:value="attachmentService.uploadProgress"
			class="progress is-primary"
			max="100"
			v-if="attachmentService.uploadProgress > 0">
			{{ attachmentService.uploadProgress }}%
		</progress>

		<table>
			<tr>
				<th>Name</th>
				<th>Size</th>
				<th>Type</th>
				<th>Date</th>
				<th>Created&nbsp;By</th>
				<th>Action</th>
			</tr>
			<tr :key="a.id" class="attachment" v-for="a in attachments">
				<td>
					{{ a.file.name }}
				</td>
				<td>{{ a.file.getHumanSize() }}</td>
				<td>{{ a.file.mime }}</td>
				<td v-tooltip="formatDate(a.created)">{{ formatDateSince(a.created) }}</td>
				<td class="has-text-centered">
					<user :avatar-size="30" :user="a.createdBy" :show-username="false" :is-inline="true"/>
				</td>
				<td>
					<div class="buttons has-addons">
						<a
							@click="downloadAttachment(a)"
							class="button is-primary noshadow"
							v-tooltip="'Download this attachment'">
							<span class="icon">
								<icon icon="cloud-download-alt"/>
							</span>
						</a>
						<a
							@click="() => {attachmentToDelete = a; showDeleteModal = true}"
							class="button is-danger noshadow" v-if="editEnabled"
							v-tooltip="'Delete this attachment'">
							<span class="icon">
								<icon icon="trash-alt"/>
							</span>
						</a>
					</div>
				</td>
			</tr>
		</table>

		<!-- Dropzone -->
		<div :class="{ 'hidden': !showDropzone }" class="dropzone" v-if="editEnabled">
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
			@close="showDeleteModal = false"
			v-if="showDeleteModal"
			@submit="deleteAttachment()">
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
import {mapState} from 'vuex'

export default {
	name: 'attachments',
	components: {
		User,
	},
	data() {
		return {
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
		},
		editEnabled: {
			default: true,
		},
	},
	created() {
		this.attachmentService = new AttachmentService()
	},
	computed: mapState({
		attachments: state => state.attachments.attachments,
	}),
	mounted() {
		document.addEventListener('dragenter', e => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = true
		})

		window.addEventListener('dragleave', e => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = false
		})

		document.addEventListener('dragover', e => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = true
		})

		document.addEventListener('drop', e => {
			e.stopPropagation()
			e.preventDefault()

			let files = e.dataTransfer.files
			this.uploadFiles(files)
			this.showDropzone = false
		})
	},
	methods: {
		downloadAttachment(attachment) {
			this.attachmentService.download(attachment)
		},
		uploadNewAttachment() {
			if (this.$refs.files.files.length === 0) {
				return
			}

			this.uploadFiles(this.$refs.files.files)
		},
		uploadFiles(files) {
			const attachmentModel = new AttachmentModel({taskId: this.taskId})
			this.attachmentService.create(attachmentModel, files)
				.then(r => {
					if (r.success !== null) {
						r.success.forEach(a => {
							this.$store.commit('attachments/add', a)
							this.$store.dispatch('tasks/addTaskAttachment', {taskId: this.taskId, attachment: a})
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
		deleteAttachment() {
			this.attachmentService.delete(this.attachmentToDelete)
				.then(r => {
					this.$store.commit('attachments/removeById', this.attachmentToDelete.id)
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
