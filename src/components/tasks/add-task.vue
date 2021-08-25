<template>
	<div class="task-add">
		<div class="field is-grouped">
			<p class="control has-icons-left is-expanded">
				<input
					:disabled="taskService.loading"
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
					:disabled="newTaskTitle === '' || taskService.loading"
					@click="addTask()"
					icon="plus"
					:loading="taskService.loading"
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
import TaskService from '../../services/task'
import createTask from '@/components/tasks/mixins/createTask'
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'

export default {
	name: 'add-task',
	data() {
		return {
			newTaskTitle: '',
			taskService: TaskService,
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
		this.taskService = new TaskService()
	},
	props: {
		defaultPosition: {
			type: Number,
			required: false,
		},
	},
	methods: {
		addTask() {
			if (this.newTaskTitle === '') {
				this.errorMessage = this.$t('list.create.addTitleRequired')
				return
			}
			this.errorMessage = ''

			if (this.taskService.loading) {
				return
			}

			this.createNewTask(this.newTaskTitle, 0, this.$store.state.auth.settings.defaultListId, this.defaultPosition)
				.then(task => {
					this.newTaskTitle = ''
					this.$emit('taskAdded', task)
				})
				.catch(e => {
					if (e === 'NO_LIST') {
						this.errorMessage = this.$t('list.create.addListRequired')
						return
					}
					this.$message.error(e)
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
