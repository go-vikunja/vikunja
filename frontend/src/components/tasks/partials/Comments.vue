<template>
	<div
		v-if="enabled"
		ref="commentsRef"
		class="content details comments-container"
	>
		<h2
			v-if="canWrite || comments.length > 0"
			class="comments-heading task-section-title"
			:class="{'d-print-none': comments.length === 0}"
		>
			<span>
				<span class="icon is-grey">
					<Icon :icon="['far', 'comments']" />
				</span>
				{{ $t('task.comment.title') }}
			</span>
			<BaseButton
				v-if="comments.length > 0"
				class="comment-sort-button"
				@click="toggleSortOrder"
			>
				<Icon :icon="commentSortOrder === 'asc' ? 'arrow-down-short-wide' : 'arrow-up-short-wide'" />
				{{ commentSortOrder === 'asc' ? $t('task.comment.sortOldestFirst') : $t('task.comment.sortNewestFirst') }}
			</BaseButton>
		</h2>
		<div class="comments">
			<span
				v-if="listPending"
				class="is-flex is-align-items-center mbs-4 mbe-4 mis-2"
			>
				<span class="loader is-inline-block mie-2" />
				{{ $t('task.comment.loading') }}
			</span>
			<div
				v-for="c in comments"
				:id="`comment-${c.id}`"
				:key="c.id"
				class="media comment"
			>
				<figure class="media-left is-hidden-mobile">
					<UserAvatar
						:user="c.author"
						:size="48"
						class="image is-avatar"
					/>
					<figcaption class="is-sr-only">
						{{ $t('misc.avatarOfUser', {user: getDisplayName(c.author)}) }}
					</figcaption>
				</figure>
				<div class="media-content">
					<div class="comment-info">
						<UserAvatar
							:user="c.author"
							:size="20"
							class="image is-avatar d-print-none"
						/>
						<strong>{{ getDisplayName(c.author) }}</strong>
						<span
							v-tooltip="formatDateLong(c.created)"
							class="has-text-grey"
						>
							{{ formatDisplayDate(c.created) }}
						</span>
						<span
							v-if="c.created && c.updated && +new Date(c.created) !== +new Date(c.updated)"
							v-tooltip="formatDateLong(c.updated)"
						>
							· {{ $t('task.comment.edited', {date: formatDisplayDate(c.updated)}) }}
						</span>
						<a
							v-tooltip="$t('task.comment.permalink')"
							:href="`#comment-${c.id}`"
							class="comment-permalink"
							:title="$t('task.comment.permalink')"
							@click.prevent.stop="copy(getCommentUrl(`${c.id}`))"
						>
							<span class="is-sr-only">{{ $t('task.comment.permalink') }}</span>
							<Icon icon="link" />
						</a>
						<CustomTransition name="fade">
							<span
								v-if="
									loading &&
										saving === c.id
								"
								class="is-inline-flex"
							>
								<span class="loader is-inline-block mie-2" />
								{{ $t('misc.saving') }}
							</span>
							<span
								v-else-if="
									!loading &&
										saved === c.id
								"
								class="has-text-success"
							>
								{{ $t('misc.saved') }}
							</span>
						</CustomTransition>
					</div>
					<Editor
						:model-value="commentDrafts[c.id] ?? c.comment"
						:is-edit-enabled="canWrite && c.author?.id === currentUserId"
						:upload-callback="attachmentUpload"
						:upload-enabled="true"
						:bottom-actions="actions[c.id]"
						:show-save="true"
						:enable-discard-shortcut="true"
						:enable-mentions="true"
						:project-id="projectId"
						initial-mode="preview"
						@update:modelValue="value => changeComment(c, value)"
						@save="() => {
							captureEditDraft(c)
							editComment(taskId)
						}"
					/>
					<Reactions 
						:model-value="c.reactions"
						class="mbs-2 d-print-none"
						entity-kind="comments"
						:entity-id="c.id"
						:task-id="taskId"
						:disabled="!canWrite"
					/>
				</div>
			</div>

			<PaginationEmit
				v-if="totalPages > 1"
				:total-pages="totalPages"
				:current-page="currentPage"
				@pageChanged="changePage"
			/>

			<div
				v-if="canWrite"
				class="media comment d-print-none"
				:class="{'new-comment-top': commentSortOrder === 'desc'}"
			>
				<figure class="media-left is-hidden-mobile">
					<UserAvatar
						:user="authStore.info"
						:size="48"
						class="image is-avatar"
					/>
					<figcaption class="is-sr-only">
						{{ $t('misc.avatarOfUser', {user: getDisplayName(authStore.info)}) }}
					</figcaption>
				</figure>
				<div class="media-content">
					<div class="form">
						<CustomTransition name="fade">
							<span
								v-if="loading && creating"
								class="is-inline-flex"
							>
								<span class="loader is-inline-block mie-2" />
								{{ $t('task.comment.creating') }}
							</span>
						</CustomTransition>
						<div class="field">
							<Editor
								v-if="editorActive"
								:key="taskId"
								ref="newCommentEditor"
								v-model="newCommentText"
								:class="{
									'is-loading':
										loading &&
										!isCommentEdit,
								}"
								:upload-callback="attachmentUpload"
								:placeholder="$t('task.comment.placeholder')"
								:enable-mentions="true"
								:project-id="projectId"
								:storage-key="commentStorageKey"
								@save="addComment()"
							/>
						</div>
						<div class="field">
							<XButton
								:loading="loading && !isCommentEdit"
								:aria-disabled="newCommentText === '' || creating"
								@click="addComment()"
							>
								{{ $t('task.comment.comment') }}
							</XButton>
						</div>
					</div>
				</div>
			</div>
		</div>


		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="() => deleteComment(commentToDelete)"
		>
			<template #header>
				<span>{{ $t('task.comment.delete') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('task.comment.deleteText1') }}<br>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, nextTick, provide, watch, onBeforeUnmount} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import Editor from '@/components/input/AsyncEditor'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import UserAvatar from '@/components/misc/UserAvatar.vue'

import {useQuery} from '@tanstack/vue-query'
import {
	commentsQuery,
	useCreateCommentMutation,
	useUpdateCommentMutation,
	useDeleteCommentMutation,
	type CommentResponse,
} from '@/client/queries/comments'

import type {Task as ITask} from '@/client/generated'

import {generateAttachmentUrl} from '@/helpers/attachments'
import {useUploadAttachmentsMutation} from '@/client/queries/attachments'
import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {clearEditorDraft} from '@/helpers/editorDraftStorage'
import {getDisplayName} from '@/helpers/user'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'
import Reactions from '@/components/input/Reactions.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {commentReplyContextKey, scrollAndHighlightComment} from '@/components/tasks/partials/commentReplyContext'

const props = withDefaults(defineProps<{
	taskId: number,
	projectId: number,
	canWrite?: boolean,
}>(), {
	canWrite: true,
})

const copy = useCopyToClipboard()

const {t} = useI18n({useScope: 'global'})
const configStore = useConfigStore()
const authStore = useAuthStore()

const localSortOrder = ref<'asc' | 'desc' | null>(null)
const commentSortOrder = computed(() => localSortOrder.value ?? authStore.settings.frontendSettings.commentSortOrder ?? 'asc')

const currentPage = ref(1)
const commentQuery = useQuery(computed(() => ({
	...commentsQuery(props.taskId, commentSortOrder.value, currentPage.value, configStore.max_items_per_page),
	enabled: configStore.task_comments_enabled && props.taskId > 0,
})))
const comments = computed(() => commentQuery.data.value?.items ?? [])
const totalPages = computed(() => commentQuery.data.value?.total_pages ?? 0)
const createMutation = useCreateCommentMutation()
const updateMutation = useUpdateCommentMutation()
const deleteMutation = useDeleteCommentMutation()
const listPending = computed(() => commentQuery.isLoading.value)
const loading = computed(() => createMutation.isPending.value || updateMutation.isPending.value)
const commentDrafts = ref<Record<number, string>>({})

const showDeleteModal = ref(false)
const commentToDelete = ref<number | null>(null)

const isCommentEdit = ref(false)
const commentEdit = ref<CommentResponse | null>(null)

const newCommentText = ref('')

const saved = ref<ITask['id'] | null>(null)
const saving = ref<ITask['id'] | null>(null)

const currentUserId = computed(() => authStore.info?.id ?? null)
const enabled = computed(() => configStore.task_comments_enabled)
const actions = computed(() => {
	if (!props.canWrite) {
		return {}
	}
	return Object.fromEntries(comments.value.map((comment) => {
		const list: {action: () => void, title: string}[] = [{
			action: () => startReplyTo(comment),
			title: t('task.comment.reply'),
		}]
		if (comment.author?.id === currentUserId.value) {
			list.push({
				action: () => toggleDelete(comment.id),
				title: t('misc.delete'),
			})
		}
		return [comment.id, list]
	}))
})

const frontend_url = computed(() => configStore.frontend_url)
const commentStorageKey = computed(() => `task-comment-${props.taskId}`)

const commentsRef = ref<HTMLElement | null>(null)
const newCommentEditor = ref<{setReplyContent: (html: string) => Promise<void>} | null>(null)

provide(commentReplyContextKey, {
	findComment: (id: number) => comments.value.find(c => c.id === id),
	scrollToComment: scrollAndHighlightComment,
})

// Strip <mention-user> elements from a reply quote so reposting the parent
// body doesn't trigger fresh notifications for users mentioned in the
// original. The inner text is kept so the quote still reads correctly.
function stripMentionsForQuote(html: string): string {
	if (!html) {
		return ''
	}
	const doc = new DOMParser().parseFromString(`<div>${html}</div>`, 'text/html')
	doc.querySelectorAll('mention-user').forEach((el) => {
		const label = (el.getAttribute('data-label') ?? el.textContent ?? '').trim()
		el.replaceWith(label ? `@${label.replace(/^@+/, '')}` : '')
	})
	return doc.body.firstElementChild?.innerHTML ?? ''
}

async function startReplyTo(parent: CommentResponse) {
	const body = stripMentionsForQuote(parent.comment ?? '')
	const draft = `<blockquote data-comment-id="${parent.id}">${body}</blockquote><p></p>`
	if (!editorActive.value) {
		editorActive.value = true
	}
	// Editor mounts asynchronously through defineAsyncComponent; wait until
	// the ref is populated before pushing content in. Bail with a warning
	// rather than fall back to `newCommentText = draft` — the modelValue
	// watcher in TipTap.vue would land the editor in preview mode, leaving
	// the user unable to type without clicking the editor first.
	const editor = await waitForEditorRef()
	if (!editor) {
		console.warn('Reply editor did not mount in time; aborting reply prefill.')
		return
	}
	await editor.setReplyContent(draft)
}

async function waitForEditorRef() {
	const start = performance.now()
	while (!newCommentEditor.value && performance.now() - start < 2000) {

		await nextTick()
	}
	return newCommentEditor.value
}


// the editor toasts the rejection itself, a mutation toast would duplicate it
const uploadAttachments = useUploadAttachmentsMutation(() => false)

function attachmentUpload(files: File[] | FileList): Promise<string[]> {
	return Promise.all(Array.from(files).map(async file => {
		const result = await uploadAttachments.mutateAsync({
			taskId: props.taskId,
			files: [file],
		})
		const [uploaded] = result.success ?? []
		// forwarded verbatim: the editor's toast translates the error code, which a rewrapped message would lose
		if (uploaded?.id === undefined) throw result.errors?.[0] ?? new Error('Attachment upload returned no file')
		return generateAttachmentUrl(props.taskId, uploaded.id)
	}))
}

async function changePage(page: number) {
	commentsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start', inline: 'nearest' })
	currentPage.value = page
}

async function toggleSortOrder() {
	const newOrder = commentSortOrder.value === 'asc' ? 'desc' : 'asc'
	if (!authStore.isLinkShareAuth) {
		await authStore.saveUserSettings({
			settings: {
				...authStore.settings,
				frontendSettings: {
					...authStore.settings.frontendSettings,
					commentSortOrder: newOrder,
					quickAddDefaultReminders: [...(authStore.settings.frontendSettings.quickAddDefaultReminders ?? [])],
				},
			},
			showMessage: false,
		})
	} else {
		localSortOrder.value = newOrder
	}
	currentPage.value = 1
}

watch(() => props.taskId, (_taskId, previousTaskId) => {
	// Reads the pending draft, so it has to run before the resets below drop it.
	flushPendingEdit(previousTaskId)
	currentPage.value = 1
	newCommentText.value = ''
	commentDrafts.value = {}
	commentEdit.value = null
	commentToDelete.value = null
	showDeleteModal.value = false
})

const editorActive = ref(true)
const creating = ref(false)

async function addComment() {
	if (!newCommentText.value || creating.value) return
	const taskId = props.taskId
	const text = newCommentText.value
	creating.value = true
	try {
		await createMutation.mutateAsync({taskId, comment: text})
		if (props.taskId !== taskId) return
		currentPage.value = commentSortOrder.value === 'desc' ? 1 : Math.max(1, totalPages.value)
		if (newCommentText.value === text) {
			newCommentText.value = ''
			clearEditorDraft(commentStorageKey.value)
		}
		if (commentSortOrder.value === 'desc') commentsRef.value?.scrollIntoView({behavior: 'smooth', block: 'start'})
	} catch {
		return
	} finally {
		creating.value = false
	}
}

function captureEditDraft(comment: CommentResponse) {
	isCommentEdit.value = true
	commentEdit.value = {...comment, comment: commentDrafts.value[comment.id] ?? comment.comment}
}

function changeComment(comment: CommentResponse, text: string) {
	commentDrafts.value[comment.id] = text
	captureEditDraft(comment)
	editCommentWithDelay()
}

function toggleDelete(id: number) {
	commentToDelete.value = id
	showDeleteModal.value = true
}

const changeTimeout = ref<ReturnType<typeof setTimeout> | null>(null)
let savedTimeout: ReturnType<typeof setTimeout> | undefined
onBeforeUnmount(() => {
	flushPendingEdit(props.taskId)
	clearTimeout(savedTimeout)
})

function cancelPendingEdit() {
	if (changeTimeout.value === null) return
	clearTimeout(changeTimeout.value)
	changeTimeout.value = null
}

function flushPendingEdit(taskId: number) {
	if (changeTimeout.value === null) return
	cancelPendingEdit()
	void editComment(taskId)
}

function editCommentWithDelay() {
	cancelPendingEdit()
	changeTimeout.value = setTimeout(() => editComment(props.taskId), 5000)
}

async function editComment(taskId: number) {
	const draft = commentEdit.value
	if (!draft?.comment) {
		isCommentEdit.value = false
		return
	}
	cancelPendingEdit()
	saving.value = draft.id
	try {
		await updateMutation.mutateAsync({taskId, id: draft.id, comment: draft.comment})
		if (props.taskId !== taskId) return
		if (commentDrafts.value[draft.id] === draft.comment) delete commentDrafts.value[draft.id]
		saved.value = draft.id
		clearTimeout(savedTimeout)
		savedTimeout = setTimeout(() => { saved.value = null }, 2000)
	} catch {
		return
	} finally {
		isCommentEdit.value = false
		saving.value = null
	}
}

async function deleteComment(id: number | null) {
	if (id === null) return
	const taskId = props.taskId
	try {
		await deleteMutation.mutateAsync({taskId, id})
	} catch {
		return
	}
	if (props.taskId === taskId && commentToDelete.value === id) {
		showDeleteModal.value = false
		commentToDelete.value = null
		currentPage.value = Math.min(currentPage.value, Math.max(1, totalPages.value))
	}
}

function getCommentUrl(commentId: string) {
	const baseUrl = frontend_url.value.endsWith('/') ? frontend_url.value.slice(0, -1) : frontend_url.value
	const url = new URL(location.pathname + location.search, baseUrl)
	url.hash = `comment-${commentId}`
	return url.toString()
}
</script>

<style lang="scss" scoped>
.media {
	align-items: flex-start;
	display: flex;
	text-align: inherit;
	padding-block-start: .5rem;

	& + .media {
		margin-block-start: .5rem;
	}
}

.media-left {
	flex-basis: auto;
	flex-grow: 0;
	flex-shrink: 0;
	margin: 0 .5rem !important;
}

.comment-info {
	display: flex;
	align-items: center;
	gap: .5rem;

	img {
		@media screen and (max-width: $tablet) {
			display: block;
			inline-size: 20px;
			block-size: 20px;
			padding-inline-end: 0;
			margin-inline-end: .5rem;
		}

		@media screen and (min-width: $tablet) {
			display: none;
		}
	}


	span,
	.comment-permalink {
		font-size: .75rem;
		line-height: 1;
	}

	.comment-permalink {
		font-size: 1rem;
		border: 1px solid transparent;
		padding: 0.25rem;
		border-radius: 1rem;
		color: var(--grey, hsl(0, 0%, 48%));
	}
	.comment-permalink:hover {
		color: var(--grey-dark, hsl(0, 0%, 29%));
		border-color: var(--grey-dark, hsl(0, 0%, 29%));
	}
}

.image.is-avatar {
	border-radius: 100%;
}

.media-content {
	flex-basis: auto;
	flex-grow: 1;
	flex-shrink: 1;
	text-align: inherit;
	inline-size: calc(100% - 48px - 2rem);
}

.comments-heading {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.comment-sort-button {
	font-size: .75rem;
	font-weight: normal;
	color: var(--grey-500);
	display: inline-flex;
	align-items: center;
	gap: .25rem;

	&:hover {
		color: var(--grey-700);
	}
}

.comments {
	display: flex;
	flex-direction: column;
}

.new-comment-top {
	order: -1;
}

.comments-container {
	scroll-margin-block-start: 4rem;
}

.media.comment {
	scroll-margin-block-start: 4rem;
	transition: background-color .3s ease-out;
	border-radius: $radius;
}

.media.comment.comment-highlight {
	background-color: hsla(var(--primary-hsl), 0.18);
	transition: background-color .15s ease-in;
}
</style>
