<template>
	<div :class="{'is-loading': taskService.loading}" class="task loader-container">
		<fancycheckbox :disabled="isArchived || disabled" @change="markAsDone" v-model="task.done"/>
		<span
			v-if="showListColor && listColor !== ''"
			:style="{backgroundColor: listColor }"
			class="color-bubble"
		>
		</span>
		<router-link
			:to="taskDetailRoute"
			:class="{ 'done': task.done}"
			class="tasktext">
			<span>
				<router-link
					:to="{ name: 'list.list', params: { listId: task.listId } }"
					class="task-list"
					v-if="showList && $store.getters['lists/getListById'](task.listId) !== null"
					v-tooltip="$t('task.detail.belongsToList', {list: $store.getters['lists/getListById'](task.listId).title})">
					{{ $store.getters['lists/getListById'](task.listId).title }}
				</router-link>

				<!-- Show any parent tasks to make it clear this task is a sub task of something -->
				<span class="parent-tasks" v-if="typeof task.relatedTasks.parenttask !== 'undefined'">
					<template v-for="(pt, i) in task.relatedTasks.parenttask">
						{{ pt.title }}<template v-if="(i + 1) < task.relatedTasks.parenttask.length">,&nbsp;</template>
					</template>
					>
				</span>
				{{ task.title }}
			</span>

			<labels class="labels ml-2 mr-1" :labels="task.labels" v-if="task.labels.length > 0"/>
			<user
				:avatar-size="27"
				:is-inline="true"
				:key="task.id + 'assignee' + a.id + i"
				:show-username="false"
				:user="a"
				v-for="(a, i) in task.assignees"
			/>
			<i
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				@click.prevent.stop="showDefer = !showDefer"
				v-if="+new Date(task.dueDate) > 0"
				v-tooltip="formatDate(task.dueDate)"
			>
				- {{ $t('task.detail.due', {at: formatDateSince(task.dueDate)}) }}
			</i>
			<transition name="fade">
				<defer-task v-if="+new Date(task.dueDate) > 0 && showDefer" v-model="task" ref="deferDueDate"/>
			</transition>
			<priority-label :priority="task.priority" :done="task.done"/>
			<span>
				<span class="list-task-icon" v-if="task.attachments.length > 0">
					<icon icon="paperclip"/>
				</span>
				<span class="list-task-icon" v-if="task.description">
					<icon icon="align-left"/>
				</span>
				<span class="list-task-icon" v-if="task.repeatAfter.amount > 0">
					<icon icon="history"/>
				</span>
			</span>
			<checklist-summary :task="task"/>
		</router-link>
		<progress
			class="progress is-small"
			v-if="task.percentDone > 0"
			:value="task.percentDone * 100" max="100">
			{{ task.percentDone * 100 }}%
		</progress>
		<router-link
			:to="{ name: 'list.list', params: { listId: task.listId } }"
			class="task-list"
			v-if="!showList && currentList.id !== task.listId && $store.getters['lists/getListById'](task.listId) !== null"
			v-tooltip="$t('task.detail.belongsToList', {list: $store.getters['lists/getListById'](task.listId).title})">
			{{ $store.getters['lists/getListById'](task.listId).title }}
		</router-link>
		<a
			:class="{'is-favorite': task.isFavorite}"
			@click="toggleFavorite"
			class="favorite">
			<icon icon="star" v-if="task.isFavorite"/>
			<icon :icon="['far', 'star']" v-else/>
		</a>
		<slot></slot>
	</div>
</template>

<script>
import TaskModel from '../../../models/task'
import PriorityLabel from './priorityLabel'
import TaskService from '../../../services/task'
import Labels from './labels'
import User from '../../misc/user'
import Fancycheckbox from '../../input/fancycheckbox'
import DeferTask from './defer-task'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {playPop} from '@/helpers/playPop'
import ChecklistSummary from './checklist-summary'

export default {
	name: 'singleTaskInList',
	data() {
		return {
			taskService: new TaskService(),
			task: new TaskModel(),
			showDefer: false,
		}
	},
	components: {
		ChecklistSummary,
		DeferTask,
		Fancycheckbox,
		User,
		Labels,
		PriorityLabel,
	},
	props: {
		theTask: {
			type: TaskModel,
			required: true,
		},
		isArchived: {
			type: Boolean,
			default: false,
		},
		showList: {
			type: Boolean,
			default: false,
		},
		disabled: {
			type: Boolean,
			default: false,
		},
		showListColor: {
			type: Boolean,
			default: true,
		},
	},
	emits: ['task-updated'],
	watch: {
		theTask(newVal) {
			this.task = newVal
		},
	},
	mounted() {
		this.task = this.theTask
		document.addEventListener('click', this.hideDeferDueDatePopup)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.hideDeferDueDatePopup)
	},
	computed: {
		listColor() {
			const list = this.$store.getters['lists/getListById'](this.task.listId)
			return list !== null ? list.hexColor : ''
		},
		currentList() {
			return typeof this.$store.state.currentList === 'undefined' ? {
				id: 0,
				title: '',
			} : this.$store.state.currentList
		},
		taskDetailRoute() {
			return {
				name: 'task.detail',
				params: { id: this.task.id },
				state: { backgroundView: this.$router.currentRoute.value.fullPath },
			}
		},
	},
	methods: {
		async markAsDone(checked) {
			const updateFunc = async () => {
				const task = await this.taskService.update(this.task)
				if (this.task.done) {
					playPop()
				}
				this.task = task
				this.$emit('task-updated', task)
				this.$message.success({
					message: this.task.done ?
						this.$t('task.doneSuccess') :
						this.$t('task.undoneSuccess'),
				}, [{
					title: 'Undo',
					callback() {
						this.task.done = !this.task.done
						this.markAsDone(!checked)
					},
				}])
			}

			if (checked) {
				setTimeout(updateFunc, 300) // Delay it to show the animation when marking a task as done
			} else {
				await updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
			}
		},

		async toggleFavorite() {
			this.task.isFavorite = !this.task.isFavorite
			this.task = await this.taskService.update(this.task)
			this.$emit('task-updated', this.task)
			this.$store.dispatch('namespaces/loadNamespacesIfFavoritesDontExist')
		},
		hideDeferDueDatePopup(e) {
			if (!this.showDefer) {
				return
			}
			closeWhenClickedOutside(e, this.$refs.deferDueDate.$el, () => {
				this.showDefer = false
			})
		},
	},
}
</script>

<style lang="scss" scoped>
.task {
	display: flex;
	flex-wrap: wrap;
	padding: .4rem;
	transition: background-color $transition;
	align-items: center;
	cursor: pointer;
	border-radius: $radius;
	border: 2px solid transparent;

	&:hover {
		background-color: var(--grey-100);
	}

	.tasktext,
	&.tasktext {
		white-space: nowrap;
		text-overflow: ellipsis;
		overflow: hidden;
		display: inline-block;
		flex: 1 0 50%;

		.overdue {
			color: var(--danger);
		}
	}

	.task-list {
		width: auto;
		color: var(--grey-400);
		font-size: .9rem;
		white-space: nowrap;
	}

	.color-bubble {
		height: 10px;
		flex: 0 0 10px;
	}

	.avatar {
		border-radius: 50%;
		vertical-align: bottom;
		margin-left: 5px;
		height: 27px;
		width: 27px;
	}

	.list-task-icon {
		margin-left: 6px;

		&:not(:first-of-type) {
			margin-left: 8px;
		}

	}

	a {
		color: var(--text);
		transition: color ease $transition-duration;

		&:hover {
			color: var(--grey-900);
		}
	}

	.favorite {
		opacity: 0;
		text-align: center;
		width: 27px;
		transition: opacity $transition, color $transition;

		&:hover {
			color: var(--warning);
		}

		&.is-favorite {
			opacity: 1;
			color: var(--warning);
		}
	}

	&:hover .favorite {
		opacity: 1;
	}

	.handle {
		opacity: 0;
		transition: opacity $transition;
		margin-right: .25rem;
		cursor: grab;
	}

	&:hover .handle {
		opacity: 1;
	}

	:deep(.fancycheckbox) {
		height: 18px;
		padding-top: 0;
		padding-right: .5rem;

		span {
			display: none;
		}
	}

	.tasktext.done {
		text-decoration: line-through;
		color: var(--grey-500);
	}

	span.parent-tasks {
		color: var(--grey-500);
		width: auto;
	}

	.remove {
		color: var(--danger);
	}

	input[type="checkbox"] {
		vertical-align: middle;
	}

	.settings {
		float: right;
		width: 24px;
		cursor: pointer;
	}

	&.loader-container.is-loading:after {
		top: calc(50% - 1rem);
		left: calc(50% - 1rem);
		width: 2rem;
		height: 2rem;
		border-left-color: var(--grey-300);
		border-bottom-color: var(--grey-300);
	}
}
</style>