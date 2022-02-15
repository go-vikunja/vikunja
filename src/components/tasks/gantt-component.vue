<template>
	<div class="gantt-chart">
		<div class="filter-container">
			<div class="items">
				<filter-popup
					v-model="params"
					@update:modelValue="loadTasks()"
				/>
			</div>
		</div>
		<div class="dates">
			<template v-for="(y, yk) in days" :key="yk + 'year'">
				<div class="months">
					<div
						:key="mk + 'month'"
						class="month"
						v-for="(m, mk) in days[yk]"
					>
						{{ formatYear(new Date(`${yk}-${parseInt(mk) + 1}-01`)) }}
						<div class="days">
							<div
								:class="{ today: d.toDateString() === now.toDateString() }"
								:key="dk + 'day'"
								:style="{ width: dayWidth + 'px' }"
								class="day"
								v-for="(d, dk) in days[yk][mk]"
							>
								<span class="theday" v-if="dayWidth > 25">
									{{ d.getDate() }}
								</span>
								<span class="weekday" v-if="dayWidth > 25">
									{{
										d.toLocaleString('en-us', {
											weekday: 'short',
										})
									}}
								</span>
							</div>
						</div>
					</div>
				</div>
			</template>
		</div>
		<div :style="{ width: fullWidth + 'px' }" class="tasks">
			<div
				v-for="(t, k) in theTasks"
				:key="t ? t.id : 0"
				:style="{
					background:
						'repeating-linear-gradient(90deg, #ededed, #ededed 1px, ' +
						(k % 2 === 0
							? '#fafafa 1px, #fafafa '
							: '#fff 1px, #fff ') +
						dayWidth +
						'px)',
				}"
				class="row"
			>
				<VueDragResize
					:class="{
						done: t ? t.done : false,
						'is-current-edit': taskToEdit !== null && taskToEdit.id === t.id,
						'has-light-text': !colorIsDark(t.getHexColor()),
						'has-dark-text': colorIsDark(t.getHexColor()),
					}"
					:gridX="dayWidth"
					:h="31"
					:isActive="canWrite"
					:minw="dayWidth"
					:parentLimitation="true"
					:parentW="fullWidth"
					:snapToGrid="true"
					:sticks="['mr', 'ml']"
					:style="{
						'border-color': t.getHexColor(),
						'background-color': t.getHexColor(),
					}"
					:w="t.durationDays * dayWidth"
					:x="t.offsetDays * dayWidth - 6"
					:y="0"
					@dragstop="(e) => resizeTask(t, e)"
					@resizestop="(e) => resizeTask(t, e)"
					axis="x"
					class="task"
				>
					<span
						:class="{
							'has-high-priority': t.priority >= priorities.HIGH,
							'has-not-so-high-priority':
								t.priority === priorities.HIGH,
							'has-super-high-priority':
								t.priority === priorities.DO_NOW,
						}"
					>
						{{ t.title }}
					</span>
					<priority-label :priority="t.priority" :done="t.done"/>
					<!-- using the key here forces vue to use the updated version model and not the response returned by the api -->
					<a @click="editTask(theTasks[k])" class="edit-toggle">
						<icon icon="pen"/>
					</a>
				</VueDragResize>
			</div>
			<template v-if="showTaskswithoutDates">
				<div
					:key="t.id"
					:style="{
						background:
							'repeating-linear-gradient(90deg, #ededed, #ededed 1px, ' +
							(k % 2 === 0
								? '#fafafa 1px, #fafafa '
								: '#fff 1px, #fff ') +
							dayWidth +
							'px)',
					}"
					class="row"
					v-for="(t, k) in tasksWithoutDates"
				>
					<VueDragResize
						:gridX="dayWidth"
						:h="31"
						:isActive="canWrite"
						:minw="dayWidth"
						:parentLimitation="true"
						:parentW="fullWidth"
						:snapToGrid="true"
						:sticks="['mr', 'ml']"
						:x="dayOffsetUntilToday * dayWidth - 6"
						:y="0"
						@dragstop="(e) => resizeTask(t, e)"
						@resizestop="(e) => resizeTask(t, e)"
						axis="x"
						class="task nodate"
						v-tooltip="$t('list.gantt.noDates')"
					>
						<span>{{ t.title }}</span>
					</VueDragResize>
				</div>
			</template>
		</div>
		<form
			@submit.prevent="addNewTask()"
			class="add-new-task"
			v-if="canWrite"
		>
			<transition name="width">
				<input
					@blur="hideCrateNewTask"
					@keyup.esc="newTaskFieldActive = false"
					class="input"
					ref="newTaskTitleField"
					type="text"
					v-if="newTaskFieldActive"
					v-model="newTaskTitle"
				/>
			</transition>
			<x-button @click="showCreateNewTask" :shadow="false" icon="plus">
				{{ $t('list.list.newTaskCta') }}
			</x-button>
		</form>
		<transition name="fade">
			<edit-task 
				v-if="isTaskEdit"
				class="taskedit"
				:title="$t('list.list.editTask')"
				@close="() => {isTaskEdit = false;taskToEdit = null}"
				:task="taskToEdit"
			/>
		</transition>
	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import VueDragResize from 'vue-drag-resize'
import EditTask from './edit-task'

import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import priorities from '../../models/constants/priorities'
import PriorityLabel from './partials/priorityLabel'
import TaskCollectionService from '../../services/taskCollection'
import {mapState} from 'vuex'
import Rights from '../../models/constants/rights.json'
import FilterPopup from '@/components/list/partials/filter-popup.vue'

import {colorIsDark} from '@/helpers/color/colorIsDark'

export default defineComponent({
	name: 'GanttChart',
	components: {
		FilterPopup,
		PriorityLabel,
		EditTask,
		VueDragResize,
	},
	props: {
		listId: {
			type: Number,
			required: true,
		},
		showTaskswithoutDates: {
			type: Boolean,
			default: false,
		},
		dateFrom: {
			default: () => new Date(new Date().setDate(new Date().getDate() - 15)),
		},
		dateTo: {
			default: () => new Date(new Date().setDate(new Date().getDate() + 30)),
		},
		// The width of a day in pixels, used to calculate all sorts of things.
		dayWidth: {
			type: Number,
			default: 35,
		},
	},
	data() {
		return {
			days: [],
			startDate: null,
			endDate: null,
			theTasks: [], // Pretty much a copy of the prop, since we cant mutate the prop directly
			tasksWithoutDates: [],
			taskService: new TaskService(),
			fullWidth: 0,
			now: new Date(),
			dayOffsetUntilToday: 0,
			isTaskEdit: false,
			taskToEdit: null,
			newTaskTitle: '',
			newTaskFieldActive: false,
			priorities: priorities,
			taskCollectionService: new TaskCollectionService(),

			params: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter_by: ['done'],
				filter_value: ['false'],
				filter_comparator: ['equals'],
				filter_concat: 'and',
			},
		}
	},
	watch: {
		dateFrom: 'buildTheGanttChart',
		dateTo: 'buildTheGanttChart',
		listId: 'parseTasks',
	},
	mounted() {
		this.buildTheGanttChart()
	},
	computed: mapState({
		canWrite: (state) => state.currentList.maxRight > Rights.READ,
	}),
	methods: {
		colorIsDark,
		buildTheGanttChart() {
			this.setDates()
			this.prepareGanttDays()
			this.parseTasks()
		},
		setDates() {
			this.startDate = new Date(this.dateFrom)
			this.endDate = new Date(this.dateTo)
			console.debug('setDates; start date: ', this.startDate, 'end date:', this.endDate, 'date from:', this.dateFrom, 'date to:', this.dateTo)

			this.dayOffsetUntilToday = Math.floor((this.now - this.startDate) / 1000 / 60 / 60 / 24) + 1
		},
		prepareGanttDays() {
			console.debug('prepareGanttDays; start date: ', this.startDate, 'end date:', this.endDate)
			// Layout: years => [months => [days]]
			let years = {}
			for (
				let d = this.startDate;
				d <= this.endDate;
				d.setDate(d.getDate() + 1)
			) {
				let date = new Date(d)
				if (years[date.getFullYear() + ''] === undefined) {
					years[date.getFullYear() + ''] = {}
				}
				if (years[date.getFullYear() + ''][date.getMonth() + ''] === undefined) {
					years[date.getFullYear() + ''][date.getMonth() + ''] = []
				}
				years[date.getFullYear() + ''][date.getMonth() + ''].push(date)
				this.fullWidth += this.dayWidth
			}
			console.debug('prepareGanttDays; years:', years)
			this.days = years
		},

		parseTasks() {
			this.setDates()
			this.loadTasks()
		},

		async loadTasks() {
			this.theTasks = []
			this.tasksWithoutDates = []

			const getAllTasks = async (page = 1) => {
				const tasks = await this.taskCollectionService.getAll({listId: this.listId}, this.params, page)
				if (page < this.taskCollectionService.totalPages) {
					const nextTasks = await getAllTasks(page + 1)
					return tasks.concat(nextTasks)
				}
				return tasks
			}

			const tasks = await getAllTasks()
			this.theTasks = tasks
				.filter((t) => {
					if (t.startDate === null && !t.done) {
						this.tasksWithoutDates.push(t)
					}
					return (
						t.startDate >= this.startDate &&
						t.endDate <= this.endDate
					)
				})
				.map((t) => this.addGantAttributes(t))
				.sort(function (a, b) {
					if (a.startDate < b.startDate) return -1
					if (a.startDate > b.startDate) return 1
					return 0
				})
		},
		addGantAttributes(t) {
			if (typeof t.durationDays !== 'undefined' && typeof t.offsetDays !== 'undefined') {
				return t
			}

			t.endDate === null ? this.endDate : t.endDate
			t.durationDays = Math.floor((t.endDate - t.startDate) / 1000 / 60 / 60 / 24)
			t.offsetDays = Math.floor((t.startDate - this.startDate) / 1000 / 60 / 60 / 24)
			return t
		},
		async resizeTask(taskDragged, newRect) {
			if (this.isTaskEdit) {
				return
			}

			let newTask = {...taskDragged}

			const didntHaveDates = newTask.startDate === null ? true : false

			let startDate = new Date(this.startDate)
			startDate.setDate(
				startDate.getDate() + newRect.left / this.dayWidth,
			)
			startDate.setUTCHours(0)
			startDate.setUTCMinutes(0)
			startDate.setUTCSeconds(0)
			startDate.setUTCMilliseconds(0)
			newTask.startDate = startDate
			let endDate = new Date(startDate)
			endDate.setDate(
				startDate.getDate() + newRect.width / this.dayWidth,
			)
			newTask.startDate = startDate
			newTask.endDate = endDate

			// We take the task from the overall tasks array because the one in it has bad data after it was updated once.
			// FIXME: This is a workaround. We should use a better mechanism to get the task or, even better,
			// prevent it from containing outdated Data in the first place.
			for (const tt in this.theTasks) {
				if (this.theTasks[tt].id === newTask.id) {
					newTask = this.theTasks[tt]
					break
				}
			}

			const ganttData = {
				endDate: newTask.endDate,
				durationDays: newTask.durationDays,
				offsetDays: newTask.offsetDays,
			}

			const r = await this.taskService.update(newTask)
			r.endDate = ganttData.endDate
			r.durationDays = ganttData.durationDays
			r.offsetDays = ganttData.offsetDays

			// If the task didn't have dates before, we'll update the list
			if (didntHaveDates) {
				for (const t in this.tasksWithoutDates) {
					if (this.tasksWithoutDates[t].id === r.id) {
						this.tasksWithoutDates.splice(t, 1)
						break
					}
				}
				this.theTasks.push(this.addGantAttributes(r))
			} else {
				for (const tt in this.theTasks) {
					if (this.theTasks[tt].id === r.id) {
						this.theTasks[tt] = this.addGantAttributes(r)
						break
					}
				}
			}
		},
		editTask(task) {
			this.taskToEdit = task
			this.isTaskEdit = true
		},
		showCreateNewTask() {
			if (!this.newTaskFieldActive) {
				// Timeout to not send the form if the field isn't even shown
				setTimeout(() => {
					this.newTaskFieldActive = true
					this.$nextTick(() => this.$refs.newTaskTitleField.focus())
				}, 100)
			}
		},
		hideCrateNewTask() {
			if (this.newTaskTitle === '') {
				this.$nextTick(() => (this.newTaskFieldActive = false))
			}
		},
		async addNewTask() {
			if (!this.newTaskFieldActive) {
				return
			}
			let task = new TaskModel({
				title: this.newTaskTitle,
				listId: this.listId,
			})
			const r = await this.taskService.create(task)
			this.tasksWithoutDates.push(this.addGantAttributes(r))
			this.newTaskTitle = ''
			this.hideCrateNewTask()
		},
		formatYear(date) {
			return this.format(date, 'MMMM, yyyy')
		},
	},
})
</script>

<style lang="scss" scoped>
$gantt-border: 1px solid var(--grey-200);
$gantt-vertical-border-color: var(--grey-100);

.gantt-chart {
	overflow-x: auto;
	border-top: 1px solid var(--grey-200);

	.dates {
		display: flex;
		text-align: center;

		.months {
			display: flex;

			.month {
				padding: 0.5rem 0 0;
				border-right: $gantt-border;
				font-family: $vikunja-font;
				font-weight: bold;

				&:last-child {
					border-right: none;
				}

				.days {
					display: flex;

					.day {
						padding: 0.5rem 0;
						font-weight: normal;

						&.today {
							background: var(--primary);
							color: var(--white);
							border-radius: 5px 5px 0 0;
							font-weight: bold;
						}

						.theday {
							padding: 0 .5rem;
							width: 100%;
							display: block;
						}

						.weekday {
							font-size: 0.8rem;
						}
					}
				}
			}
		}
	}

	.tasks {
		max-width: unset !important;
		border-top: $gantt-border;

		.row {
			height: 45px;

			.task {
				display: inline-block;
				border: 2px solid var(--primary);
				font-size: 0.85rem;
				margin: 0.5rem;
				border-radius: 6px;
				padding: 0.25rem 0.5rem;
				cursor: grab;
				position: relative;
				height: 31px !important;

				-webkit-touch-callout: none; // iOS Safari
				user-select: none; // Non-prefixed version

				&.is-current-edit {
					border-color: var(--warning) !important;
				}

				&.has-light-text {
					color: var(--light);

					&.done span:after {
						border-top: 1px solid var(--light);
					}

					.edit-toggle {
						color: var(--light);
					}
				}

				&.has-dark-text {
					color: var(--text);

					&.done span:after {
						border-top: 1px solid var(--dark);
					}

					.edit-toggle {
						color: var(--text);
					}
				}

				&.done span {
					position: relative;

					&::after {
						content: '';
						position: absolute;
						right: 0;
						left: 0;
						top: 57%;
					}
				}

				span:not(.high-priority) {
					max-width: calc(100% - 20px);
					display: inline-block;
					white-space: nowrap;
					text-overflow: ellipsis;
					overflow: hidden;

					&.has-high-priority {
						max-width: calc(100% - 90px);
					}

					&.has-not-so-high-priority {
						max-width: calc(100% - 70px);
					}

					&.has-super-high-priority {
						max-width: calc(100% - 111px);
					}

					&.icon {
						width: 10px;
						text-align: center;
					}
				}

				.high-priority {
					margin: 0 0 0 .5rem;
					vertical-align: bottom;
				}

				.edit-toggle {
					float: right;
					cursor: pointer;
					margin-right: 4px;
				}

				&.nodate {
					border: 2px dashed var(--grey-300);
					background: var(--grey-100);
				}

				&:active {
					cursor: grabbing;
				}
			}
		}
	}

	.taskedit {
		position: fixed;
		top: 10vh;
		right: 10vw;
		z-index: 5;

		// FIXME: should be an option of the card, e.g. overflow
		:deep(.card-content) {
			max-height: 60vh;
			overflow-y: auto;
		}
	}

	.add-new-task {
		padding: 1rem .7rem .4rem .7rem;
		display: flex;
		max-width: 450px;

		.input {
			margin-right: .7rem;
			font-size: .8rem;
		}

		.button {
			font-size: .68rem;
		}
	}
}
</style>