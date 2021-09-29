<template>
	<div class="task-add">
		<div class="field is-grouped">
			<p class="control has-icons-left is-expanded">
				<textarea
					:disabled="taskService.loading || null"
					class="input"
					:placeholder="$t('list.list.addPlaceholder')"
					type="text"
					v-focus
					v-model="newTaskTitle"
					ref="newTaskInput"
					:style="{'height': `${textAreaHeight}px`}"
					@keyup="errorMessage = ''"
					@keydown.enter="handleEnter"
				/>
				<span class="icon is-small is-left">
					<icon icon="tasks"/>
				</span>
			</p>
			<p class="control">
				<x-button
					:disabled="newTaskTitle === '' || taskService.loading || null"
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

const INITIAL_SCROLL_HEIGHT = 40

const cleanupTitle = title => {
	return title.replace(/^((\* |\+ |- )(\[ \] )?)/g, '')
}

export default {
	name: 'add-task',
	data() {
		return {
			newTaskTitle: '',
			taskService: new TaskService(),
			errorMessage: '',
			textAreaHeight: INITIAL_SCROLL_HEIGHT,
		}
	},
	mixins: [
		createTask,
	],
	components: {
		QuickAddMagic,
	},
	props: {
		defaultPosition: {
			type: Number,
			required: false,
		},
	},
	watch: {
		newTaskTitle(newVal) {
			let scrollHeight = this.$refs.newTaskInput.scrollHeight
			if (scrollHeight < INITIAL_SCROLL_HEIGHT || newVal === '') {
				scrollHeight = INITIAL_SCROLL_HEIGHT
			}

			this.textAreaHeight = scrollHeight
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

			const newTasks = []
			this.newTaskTitle.split(/[\r\n]+/).forEach(t => {
				const title = cleanupTitle(t)
				if (title === '') {
					return
				}
				
				newTasks.push(
					this.createNewTask(title, 0, this.$store.state.auth.settings.defaultListId, this.defaultPosition)
						.then(task => {
							this.$emit('taskAdded', task)
							return task
						}),
				)
			})

			Promise.all(newTasks)
				.then(() => {
					this.newTaskTitle = ''
				})
				.catch(e => {
					if (e === 'NO_LIST') {
						this.errorMessage = this.$t('list.create.addListRequired')
						return
					}
					this.$message.error(e)
				})
		},
		handleEnter(e) {
			// when pressing shift + enter we want to continue as we normally would. Otherwise, we want to create 
			// the new task(s). The vue event modifier don't allow this, hence this method.
			if (e.shiftKey) {
				return
			}

			e.preventDefault()
			this.addTask()
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

.input, .textarea {
	transition: border-color $transition;
}
</style>
