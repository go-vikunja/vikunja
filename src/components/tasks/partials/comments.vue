<template>
	<div :class="{'has-top-border': canWrite || comments.length > 0}" class="content details">
		<h1 v-if="canWrite || comments.length > 0">
			<span class="icon is-grey">
				<icon :icon="['far', 'comments']"/>
			</span>
			Comments
		</h1>
		<div class="comments">
			<progress class="progress is-small is-info" max="100" v-if="taskCommentService.loading">Loading
				comments...
			</progress>
			<div :key="c.id" class="media comment" v-for="c in comments">
				<figure class="media-left">
					<img :src="c.author.getAvatarUrl(48)" alt="" class="image is-avatar" height="48" width="48"/>
				</figure>
				<div class="media-content">
					<div :class="{'is-pulled-up': canWrite}" class="comment-info">
						<strong>{{ c.author.username }}</strong>&nbsp;
						<small v-tooltip="formatDate(c.created)">{{ formatDateSince(c.created) }}</small>
						<small v-if="+new Date(c.created) !== +new Date(c.updated)" v-tooltip="formatDate(c.updated)"> ·
							edited {{ formatDateSince(c.updated) }}</small>
					</div>
					<editor
						:has-preview="true"
						:is-edit-enabled="canWrite"
						:upload-callback="attachmentUpload"
						:upload-enabled="true"
						@change="() => {toggleEdit(c);editComment()}"
						v-model="c.comment"
					/>
					<div class="comment-actions" v-if="canWrite">
						<a @click="toggleDelete(c.id)">Remove</a>
					</div>
				</div>
			</div>
			<div class="media comment" v-if="canWrite">
				<figure class="media-left">
					<img :src="userAvatar" alt="" class="image is-avatar" height="48" width="48"/>
				</figure>
				<div class="media-content">
					<div class="form">
						<div class="field">
							<editor
								:class="{'is-loading': taskCommentService.loading && !isCommentEdit}"
								:has-preview="false"
								:upload-callback="attachmentUpload"
								:upload-enabled="true"
								placeholder="Add your comment..."
								v-if="editorActive"
								v-model="newComment.comment"
							/>
						</div>
						<div class="field">
							<button :class="{'is-loading': taskCommentService.loading && !isCommentEdit}"
									:disabled="newComment.comment === ''"
									@click="addComment()" class="button is-primary">Comment
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<modal
			@close="showDeleteModal = false"
			@submit="deleteComment()"
			v-if="showDeleteModal">
			<span slot="header">Delete this comment</span>
			<p slot="text">Are you sure you want to delete this comment?
				<br/>This <b>CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
import TaskCommentService from '../../../services/taskComment'
import TaskCommentModel from '../../../models/taskComment'
import attachmentUpload from '../mixins/attachmentUpload'
import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'

export default {
	name: 'comments',
	components: {
		editor: () => ({
			component: import(/* webpackPrefetch: true *//* webpackChunkName: "editor" */ '../../input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	mixins: [
		attachmentUpload,
	],
	props: {
		taskId: {
			type: Number,
			required: true,
		},
		canWrite: {
			default: true,
		},
	},
	data() {
		return {
			comments: [],

			showDeleteModal: false,
			commentToDelete: TaskCommentModel,

			isCommentEdit: false,
			commentEdit: TaskCommentModel,

			taskCommentService: TaskCommentService,
			newComment: TaskCommentModel,
			editorActive: true,
		}
	},
	created() {
		this.taskCommentService = new TaskCommentService()
		this.newComment = new TaskCommentModel({taskId: this.taskId})
		this.commentEdit = new TaskCommentModel({taskId: this.taskId})
		this.commentToDelete = new TaskCommentModel({taskId: this.taskId})
		this.comments = []
	},
	mounted() {
		this.loadComments()
	},
	watch: {
		taskId() {
			this.loadComments()
		},
	},
	computed: {
		userAvatar() {
			return this.$store.state.auth.info.getAvatarUrl(48)
		},
	},
	methods: {
		loadComments() {
			this.taskCommentService.getAll({taskId: this.taskId})
				.then(r => {
					this.$set(this, 'comments', r)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		addComment() {
			if (this.newComment.comment === '') {
				return
			}

			// This makes the editor trigger its mounted function again which makes it forget every input
			// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
			// which made it impossible to detect change from the outside. Therefore the component would
			// not update if new content from the outside was made available.
			// See https://github.com/NikulinIlya/vue-easymde/issues/3
			this.editorActive = false
			this.$nextTick(() => this.editorActive = true)

			this.taskCommentService.create(this.newComment)
				.then(r => {
					this.comments.push(r)
					this.newComment.comment = ''
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		toggleEdit(comment) {
			this.isCommentEdit = !this.isCommentEdit
			this.commentEdit = comment
		},
		toggleDelete(commentId) {
			this.showDeleteModal = !this.showDeleteModal
			this.commentToDelete.id = commentId
		},
		editComment() {
			if (this.commentEdit.comment === '') {
				return
			}
			this.commentEdit.taskId = this.taskId
			this.taskCommentService.update(this.commentEdit)
				.then(r => {
					for (const c in this.comments) {
						if (this.comments[c].id === this.commentEdit.id) {
							this.$set(this.comments, c, r)
						}
					}
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.isCommentEdit = false
				})
		},
		deleteComment() {
			this.taskCommentService.delete(this.commentToDelete)
				.then(() => {
					for (const a in this.comments) {
						if (this.comments[a].id === this.commentToDelete.id) {
							this.comments.splice(a, 1)
						}
					}
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
