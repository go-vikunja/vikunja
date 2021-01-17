<template>
	<div :class="{'is-loading': taskService.loading}" class="task loader-container">
		<fancycheckbox :disabled="isArchived || disabled" @change="markAsDone" v-model="task.done"/>
		<span
			:style="{backgroundColor: listColor }"
			class="color-bubble"
			v-if="listColor !== ''">
		</span>
		<router-link
			:to="{ name: taskDetailRoute, params: { id: task.id } }"
			:class="{ 'done': task.done}"
			class="tasktext">
			<span>
				<router-link
					:to="{ name: 'list.list', params: { listId: task.listId } }"
					class="task-list"
					v-if="showList && $store.getters['lists/getListById'](task.listId) !== null"
					v-tooltip="`This task belongs to list '${$store.getters['lists/getListById'](task.listId).title}'`">
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

			<labels :labels="task.labels"/>
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
				- Due {{ formatDateSince(task.dueDate) }}
			</i>
			<transition name="fade">
				<defer-task v-if="+new Date(task.dueDate) > 0 && showDefer" v-model="task" ref="deferDueDate"/>
			</transition>
			<priority-label :priority="task.priority"/>
			<span>
				<span class="list-task-icon" v-if="task.attachments.length > 0">
					<icon icon="paperclip"/>
				</span>
				<span class="list-task-icon" v-if="task.description">
					<icon icon="align-left"/>
				</span>
			</span>
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
			v-tooltip="`This task belongs to list '${$store.getters['lists/getListById'](task.listId).title}'`">
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

export default {
	name: 'singleTaskInList',
	data() {
		return {
			taskService: TaskService,
			task: TaskModel,
			showDefer: false,
		}
	},
	components: {
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
		taskDetailRoute: {
			type: String,
			default: 'task.list.detail',
		},
		showList: {
			type: Boolean,
			default: false,
		},
		disabled: {
			type: Boolean,
			default: false,
		},
	},
	watch: {
		theTask(newVal) {
			this.task = newVal
		},
	},
	mounted() {
		this.task = this.theTask
		document.addEventListener('click', this.hideDeferDueDatePopup)
	},
	created() {
		this.task = new TaskModel()
		this.taskService = new TaskService()
	},
	beforeDestroy() {
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
	},
	methods: {
		markAsDone(checked) {
			const updateFunc = () => {
				this.taskService.update(this.task)
					.then(t => {
						this.task = t
						this.$emit('task-updated', t)
						this.success(
							{message: 'The task was successfully ' + (this.task.done ? '' : 'un-') + 'marked as done.'},
							this,
							[{
								title: 'Undo',
								callback: () => this.markAsDone({
									target: {
										checked: !checked,
									},
								}),
							}],
						)
					})
					.catch(e => {
						this.error(e, this)
					})
			}

			if (checked) {
				setTimeout(updateFunc, 300) // Delay it to show the animation when marking a task as done
			} else {
				updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
			}
		},
		toggleFavorite() {
			this.task.isFavorite = !this.task.isFavorite
			this.taskService.update(this.task)
				.then(t => {
					this.task = t
					this.$emit('task-updated', t)
					this.$store.dispatch('namespaces/loadNamespacesIfFavoritesDontExist')
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		hideDeferDueDatePopup(e) {
			if (this.showDefer) {
				closeWhenClickedOutside(e, this.$refs.deferDueDate.$el, () => {
					this.showDefer = false
				})
			}
		},
	},
}
</script>
