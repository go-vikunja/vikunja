<template>
	<div class="task-add">
		<div class="field is-grouped">
			<p :class="{ 'is-loading': taskService.loading}" class="control has-icons-left is-expanded">
				<input
					:class="{ 'disabled': taskService.loading}"
					@keyup.enter="addTask()"
					class="input"
					:placeholder="$t('list.list.addPlaceholder')"
					type="text"
					v-focus
					v-model="newTaskTitle"
					ref="newTaskInput"
					@keyup="errorMessage = ''"
				/>
				<span class="icon is-small is-left">
					<icon icon="tasks"/>
				</span>
			</p>
			<p class="control">
				<x-button
					:disabled="newTaskTitle.length === 0"
					@click="addTask()"
					icon="plus"
				>
					{{ $t('list.list.add') }}
				</x-button>
			</p>
		</div>
		<p class="help is-danger" v-if="errorMessage !== ''">
			{{ errorMessage }}
		</p>
		<quick-add-magic v-if="errorMessage === ''"/>
	</div>
</template>

<script>
import ListService from '../../services/list'
import TaskService from '../../services/task'
import LabelService from '../../services/label'
import LabelTaskService from '../../services/labelTask'
import createTask from '@/components/tasks/mixins/createTask'
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic'

export default {
	name: 'add-task',
	data() {
		return {
			newTaskTitle: '',
			listService: ListService,
			taskService: TaskService,
			labelService: LabelService,
			labelTaskService: LabelTaskService,
			errorMessage: '',
		}
	},
	mixins: [
		createTask,
	],
	components: {
		QuickAddMagic,
	},
	created() {
		this.listService = new ListService()
		this.taskService = new TaskService()
		this.labelService = new LabelService()
		this.labelTaskService = new LabelTaskService()
	},
	methods: {
		addTask() {
			if (this.newTaskTitle === '') {
				this.errorMessage = this.$t('list.create.addTitleRequired')
				return
			}
			this.errorMessage = ''

			this.createNewTask(this.newTaskTitle, 0, this.$store.state.auth.settings.defaultListId)
				.then(task => {
					this.newTaskTitle = ''
					this.$emit('taskAdded', task)
				})
				.catch(e => {
					if (e === 'NO_LIST') {
						this.errorMessage = this.$t('list.create.addListRequired')
						return
					}
					this.error(e)
				})
		},
	},
}
</script>

<style lang="scss" scoped>
.task-add {
	margin-bottom: 0;

	.button {
		height: 2.5rem;
	}
}
</style>
