<template>
	<div
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': !colorIsDark(task.hexColor) && task.hexColor !== `#${task.defaultColor}` && task.hexColor !== task.defaultColor,
		}"
		:style="{'background-color': task.hexColor !== '#' && task.hexColor !== `#${task.defaultColor}` ? task.hexColor : false}"
		@click.ctrl="() => toggleTaskDone(task)"
		@click.exact="() => $router.push({ name: 'task.kanban.detail', params: { id: task.id } })"
		@click.meta="() => toggleTaskDone(task)"
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
		async toggleTaskDone(task) {
			this.loadingInternal = true
			try {
				await this.$store.dispatch('tasks/update', {
					...task,
					done: !task.done,
				})
				if (task.done) {
					playPop()
				}
			} finally {
				this.loadingInternal = false
			}
		},
	},
}
</script>

<style lang="scss" scoped>
$task-background: $white;

.task {
	-webkit-touch-callout: none; // iOS Safari
	user-select: none;
	cursor: pointer;
	box-shadow: $shadow-xs;
	display: block;
	border: 3px solid transparent;

	font-size: .9rem;
	margin: .5rem;
	padding: .4rem;
	border-radius: $radius;
	background: $task-background;

	&.loader-container.is-loading::after {
		width: 1.5rem;
		height: 1.5rem;
		top: calc(50% - .75rem);
		left: calc(50% - .75rem);
		border-width: 2px;
	}

	h3 {
		font-family: $family-sans-serif;
		font-size: .85rem;
		word-break: break-word;
	}

	.progress {
		margin: 8px 0 0 0;
		width: 100%;
		height: 0.5rem;
	}

	.due-date {
		float: right;
		display: flex;
		align-items: center;

		.icon {
			margin-right: .25rem;
		}

		&.overdue {
			color: $red;
		}
	}

	.label-wrapper .tag {
		margin: .5rem .5rem 0 0;
	}

	.footer {
		background: transparent;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		align-items: center;

		::v-deep.tag,
		.assignees,
		.icon,
		.priority-label {
			margin-top: .25rem;
			margin-right: .25rem;
		}

		.assignees {
			display: flex;

			.user {
				display: inline;
				margin: 0;

				img {
					margin: 0;
				}
			}
		}

		// FIXME: should be in labels.vue
		::v-deep.tag {
			margin-left: 0;
		}

		.priority-label {
			font-size: .75rem;
			height: 2rem;

			.icon {
				height: 1rem;
				padding: 0 .25rem;
				margin-top: 0;
			}
		}
	}

	.footer .icon,
	.due-date,
	.priority-label {
		background: $grey-100;
		border-radius: $radius;
		padding: 0 .5rem;
	}

	.due-date {
		padding: 0 .25rem;
	}

	.task-id {
		color: $grey-500;
		font-size: .8rem;
		margin-bottom: .25rem;
		display: flex;
	}

	.is-done {
		font-size: .75rem;
		padding: .2rem .3rem;
		margin: 0 .25rem 0 0;
	}

	&.is-moving {
		opacity: .5;
	}

	span {
		width: auto;
	}

	&.has-light-text {
		color: $white;

		.task-id {
			color: $grey-200;
		}

		.footer .icon,
		.due-date,
		.priority-label {
			background: $grey-800;
		}

		.footer {
			.icon svg {
				fill: $white;
			}
		}
	}
}
</style>