<template>
	<div
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': !colorIsDark(task.hexColor) && task.hexColor !== `#${task.defaultColor}` && task.hexColor !== task.defaultColor,
		}"
		:style="{'background-color': task.hexColor !== '#' && task.hexColor !== `#${task.defaultColor}` ? task.hexColor : false}"
		@click.ctrl="() => markTaskAsDone(task)"
		@click.exact="() => $router.push({ name: 'task.kanban.detail', params: { id: task.id } })"
		@click.meta="() => markTaskAsDone(task)"
		class="task loader-container draggable"
	>
		<span class="task-id">
			<span class="is-done" v-if="task.done">Done</span>
			<template v-if="task.identifier === ''">
				#{{ task.index }}
			</template>
			<template v-else>
				{{ task.identifier }}
			</template>
		</span>
		<span
			:class="{'overdue': task.dueDate <= new Date() && !task.done}"
			class="due-date"
			v-if="task.dueDate > 0"
			v-tooltip="formatDate(task.dueDate)">
			<span class="icon">
				<icon :icon="['far', 'calendar-alt']"/>
			</span>
			<span>
				{{ formatDateSince(task.dueDate) }}
			</span>
		</span>
		<h3>{{ task.title }}</h3>
		<progress
			class="progress is-small"
			v-if="task.percentDone > 0"
			:value="task.percentDone * 100" max="100">
			{{ task.percentDone * 100 }}%
		</progress>
		<div class="footer">
			<labels :labels="task.labels"/>
			<priority-label :priority="task.priority" :done="task.done"/>
			<div class="assignees" v-if="task.assignees.length > 0">
				<user
					:avatar-size="24"
					:key="task.id + 'assignee' + u.id"
					:show-username="false"
					:user="u"
					v-for="u in task.assignees"
				/>
			</div>
			<checklist-summary :task="task"/>
			<span class="icon" v-if="task.attachments.length > 0">
				<icon icon="paperclip"/>	
			</span>
			<span v-if="task.description" class="icon">
				<icon icon="align-left"/>
			</span>
		</div>
	</div>
</template>

<script>
import {playPop} from '../../../helpers/playPop'
import PriorityLabel from '../../../components/tasks/partials/priorityLabel'
import User from '../../../components/misc/user'
import Labels from '../../../components/tasks/partials/labels'
import ChecklistSummary from './checklist-summary'

export default {
	name: 'kanban-card',
	components: {
		ChecklistSummary,
		PriorityLabel,
		User,
		Labels,
	},
	data() {
		return {
			loadingInternal: false,
		}
	},
	props: {
		task: {
			required: true,
		},
		loading: {
			type: Boolean,
			required: false,
			default: false,
		},
	},
	methods: {
		markTaskAsDone(task) {
			this.loadingInternal = true
			this.$store.dispatch('tasks/update', {
				...task,
				done: !task.done,
			})
				.then(() => {
					if (task.done) {
						playPop()
					}
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					this.loadingInternal = false
				})
		},
	},
}
</script>
