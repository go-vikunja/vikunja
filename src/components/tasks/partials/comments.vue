<template>
	<div class="content details has-top-border">
		<h1>
			<span class="icon is-grey">
				<icon :icon="['far', 'comments']"/>
			</span>
			Comments
		</h1>
		<div class="comments">
			<progress class="progress is-small is-info" max="100" v-if="taskCommentService.loading">Loading comments...</progress>
			<div class="media comment" v-for="c in comments" :key="c.id">
				<figure class="media-left">
					<img class="image is-avatar" :src="c.author.getAvatarUrl(48)" alt="" width="48" height="48"/>
				</figure>
				<div class="media-content">
					<div class="form" v-if="isCommentEdit && commentEdit.id === c.id">
						<div class="field">
							<textarea class="textarea" :class="{'is-loading': taskCommentService.loading}" placeholder="Add your comment..." v-model="commentEdit.comment" @keyup.ctrl.enter="editComment()"></textarea>
						</div>
						<div class="field">
							<button class="button is-primary" :class="{'is-loading': taskCommentService.loading}" @click="editComment()" :disabled="commentEdit.comment === ''">Comment</button>
							<a @click="() => isCommentEdit = false">Cancel</a>
						</div>
					</div>
					<div class="content" v-else>
						<strong>{{ c.author.username }}</strong>&nbsp;
						<small v-tooltip="formatDate(c.created)">{{ formatDateSince(c.created) }}</small>
						<small v-if="+new Date(c.created) !== +new Date(c.updated)" v-tooltip="formatDate(c.updated)"> · edited {{ formatDateSince(c.updated) }}</small>
						<br/>
						<p>
							{{c.comment}}
						</p>
						<div class="comment-actions">
							<a @click="toggleEdit(c)">Edit</a>&nbsp;·&nbsp;
							<a @click="toggleDelete(c.id)">Remove</a>
						</div>
					</div>
				</div>
			</div>
			<div class="media comment">
				<figure class="media-left">
					<img class="image is-avatar" :src="userAvatar" alt="" width="48" height="48"/>
				</figure>
				<div class="media-content">
					<div class="form">
						<div class="field">
							<textarea class="textarea" :class="{'is-loading': taskCommentService.loading && !isCommentEdit}" placeholder="Add your comment..." v-model="newComment.comment" @keyup.ctrl.enter="addComment()"></textarea>
						</div>
						<div class="field">
							<button class="button is-primary" :class="{'is-loading': taskCommentService.loading && !isCommentEdit}" @click="addComment()" :disabled="newComment.comment === ''">Comment</button>
						</div>
					</div>
				</div>
			</div>
		</div>
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteComment()">
			<span slot="header">Delete this comment</span>
			<p slot="text">Are you sure you want to delete this comment?
				<br/>This <b>CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import TaskCommentService from '../../../services/taskComment'
	import TaskCommentModel from '../../../models/taskComment'

	export default {
		name: 'comments',
		props: {
			taskId: {
				type: Number,
				required: true,
			}
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
			}
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
