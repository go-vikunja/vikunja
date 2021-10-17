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
					:style="{'height': `calc(${textAreaHeight}px - 2px + 1rem)`}"
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
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'

const INPUT_BORDER_PX = 2
const LINE_HEIGHT = 1.5 // using getComputedStyles().lineHeight returns an (wrong) absolute pixel value, we need the factor to do calculations with it.

const cleanupTitle = title => {
	return title.replace(/^((\* |\+ |- )(\[ \] )?)/g, '')
}

export default {
	name: 'add-task',
	emits: ['taskAdded'],
	data() {
		return {
			newTaskTitle: '',
			taskService: new TaskService(),
			errorMessage: '',
			textAreaHeight: null,
			initialTextAreaHeight: null,
		}
	},
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
			// Calculating the textarea height based on lines of input in it. That is more reliable when removing a 
			// line from the input.
			const numberOfLines = newVal.split(/\r\n|\r|\n/).length
			const fontSize =  parseInt(window.getComputedStyle(this.$refs.newTaskInput, null).getPropertyValue('font-size'))

			this.textAreaHeight = numberOfLines * fontSize * LINE_HEIGHT + INPUT_BORDER_PX
		},
	},
	mounted() {
		this.initialTextAreaHeight = this.$refs.newTaskInput.scrollHeight + INPUT_BORDER_PX
	},
	methods: {
		async addTask() {
			if (this.newTaskTitle === '') {
				this.errorMessage = this.$t('list.create.addTitleRequired')
				return
			}
			this.errorMessage = ''

			if (this.taskService.loading) {
				return
			}

			const newTasks = this.newTaskTitle.split(/[\r\n]+/).map(async t => {
				const title = cleanupTitle(t)
				if (title === '') {
					return
				}
				
				const task = await this.$store.dispatch('tasks/createNewTask', {
					title: this.newTaskTitle,
					listId: this.$store.state.auth.settings.defaultListId,
					position: this.defaultPosition,
				})
				this.$emit('taskAdded', task)
				return task
			})

			try {
				await Promise.all(newTasks)
				this.newTaskTitle = ''
			} catch(e) {
				if (e.message === 'NO_LIST') {
					this.errorMessage = this.$t('list.create.addListRequired')
					return
				}
				throw e
			}
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
