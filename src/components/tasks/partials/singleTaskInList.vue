<template>
	<span>
		<fancycheckbox v-model="task.done" @change="markAsDone" :disabled="isArchived"/>
		<router-link :to="{ name: taskDetailRoute, params: { id: task.id } }" class="tasktext"  :class="{ 'done': task.done}">

			<router-link
					v-if="showList && $store.getters['lists/getListById'](task.listId) !== null"
					v-tooltip="`This task belongs to list '${$store.getters['lists/getListById'](task.listId).title}'`"
					:to="{ name: 'list.list', params: { listId: task.listId } }"
					class="task-list">
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
			<labels :labels="task.labels"/>
			<user
					:user="a"
					:avatar-size="27"
					:show-username="false"
					:is-inline="true"
					v-for="(a, i) in task.assignees"
					:key="task.id + 'assignee' + a.id + i"
			/>
			<i v-if="task.dueDate > 0"
				:class="{'overdue': task.dueDate <= new Date() && !task.done}"
				v-tooltip="formatDate(task.dueDate)"> - Due {{formatDateSince(task.dueDate)}}</i>
			<priority-label :priority="task.priority"/>
		</router-link>
	</span>
</template>

<script>
	import TaskModel from '../../../models/task'
	import PriorityLabel from './priorityLabel'
	import TaskService from '../../../services/task'
	import Labels from './labels'
	import User from '../../misc/user'
	import Fancycheckbox from '../../input/fancycheckbox'

	export default {
		name: 'singleTaskInList',
		data() {
			return {
				taskService: TaskService,
				task: TaskModel,
			}
		},
		components: {
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
				default: 'task.list.detail'
			},
			showList: {
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
		},
		created() {
			this.task = new TaskModel()
			this.taskService = new TaskService()
		},
		methods: {
			markAsDone(checked) {
				const updateFunc = () => {
					this.taskService.update(this.task)
						.then(t => {
							this.task = t
							this.$emit('taskUpdated', t)
							this.success(
								{message: 'The task was successfully ' + (this.task.done ? '' : 'un-') + 'marked as done.'},
								this,
								[{
									title: 'Undo',
									callback: () => this.markAsDone({
										target: {
											checked: !checked
										}
									}),
								}]
							)
						})
						.catch(e => {
							this.error(e, this)
						})
				}

				if (checked) {
					setTimeout(updateFunc, 300); // Delay it to show the animation when marking a task as done
				} else {
					updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
				}
			},
		},
	}
</script>
