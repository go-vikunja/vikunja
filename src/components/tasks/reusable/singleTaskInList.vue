<template>
	<span>
		<div class="fancycheckbox" :class="{'is-disabled': isArchived}">
			<input @change="markAsDone" type="checkbox" :id="task.id" :checked="task.done"
				style="display: none;" :disabled="isArchived">
			<label :for="task.id" class="check">
				<svg width="18px" height="18px" viewBox="0 0 18 18">
					<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
					<polyline points="1 9 7 14 15 4"></polyline>
				</svg>
			</label>
		</div>
		<router-link :to="{ name: 'taskDetailView', params: { id: task.id } }" class="tasktext"  :class="{ 'done': task.done}">
			<!-- Show any parent tasks to make it clear this task is a sub task of something -->
			<span class="parent-tasks" v-if="typeof task.related_tasks.parenttask !== 'undefined'">
				<template v-for="(pt, i) in task.related_tasks.parenttask">
					{{ pt.text }}<template v-if="(i + 1) < task.related_tasks.parenttask.length">,&nbsp;</template>
				</template>
				>
			</span>
			{{ task.text }}
			<span class="tag" v-for="label in task.labels" :style="{'background': label.hex_color, 'color': label.textColor}"
				:key="label.id">
				<span>{{ label.title }}</span>
			</span>
			<img
					:src="a.getAvatarUrl(27)"
					:alt="a.username"
					class="avatar"
					width="27"
					height="27"
					v-for="(a, i) in task.assignees"
					:key="task.id + 'assignee' + a.id + i"/>
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

	export default {
		name: 'singleTaskInList',
		data() {
			return {
				taskService: TaskService,
				task: TaskModel,
			}
		},
		components: {
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
			markAsDone(e) {
				let updateFunc = () => {
					// We get the task, update the 'done' property and then push it to the api.
					this.task.done = e.target.checked
					let task = new TaskModel(this.task)
					task.done = e.target.checked
					this.taskService.update(task)
						.then(t => {
							this.task = t
							this.$emit('taskUpdated', t)
							this.success(
								{message: 'The task was successfully ' + (task.done ? '' : 'un-') + 'marked as done.'},
								this,
								[{
									title: 'Undo',
									callback: () => this.markAsDone({
										target: {
											id: e.target.id,
											checked: !e.target.checked
										}
									}),
								}]
							)
						})
						.catch(e => {
							this.error(e, this)
						})
				}

				if (e.target.checked) {
					setTimeout(updateFunc(), 300); // Delay it to show the animation when marking a task as done
				} else {
					updateFunc() // Don't delay it when un-marking it as it doesn't have an animation the other way around
				}
			},
		},
	}
</script>
