<template>
	<div class="heading">
		<h1 class="title task-id">
			{{ task.getTextIdentifier() }}
		</h1>
		<div class="is-done" v-if="task.done">Done</div>
		<h1
			class="title input"
			:class="{'disabled': !canWrite}"
			@focusout="save()"
			@keydown.enter.prevent.stop="save()"
			:contenteditable="canWrite ? 'true' : 'false'"
			ref="taskTitle">{{ task.title.trim() }}</h1>
		<transition name="fade">
			<span class="is-inline-flex is-align-items-center" v-if="loading && saving">
				<span class="loader is-inline-block mr-2"></span>
				Saving...
			</span>
			<span class="has-text-success is-inline-flex is-align-content-center" v-if="!loading && saved">
				<icon icon="check" class="mr-2"/>
				Saved!
			</span>
		</transition>
	</div>
</template>

<script>
import {LOADING} from '@/store/mutation-types'
import {mapState} from 'vuex'

export default {
	name: 'heading',
	data() {
		return {
			task: {title: '', identifier: '', index:''},
			taskTitle: '',
			saved: false,
			saving: false, // Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
		}
	},
	computed: mapState({
		loading: LOADING,
	}),
	props: {
		value: {
			required: true,
		},
		canWrite: {
			type: Boolean,
			default: false,
		},
	},
	watch: {
		value(newVal) {
			this.task = newVal
			this.taskTitle = this.task.title
		},
	},
	mounted() {
		this.task = this.value
		this.taskTitle = this.task.title
	},
	methods: {
		save() {
			this.$refs.taskTitle.spellcheck = false

			// Pull the task title from the contenteditable
			let taskTitle = this.$refs.taskTitle.textContent
			this.task.title = taskTitle

			// We only want to save if the title was actually change.
			// Because the contenteditable does not have a change event,
			// we're building it ourselves and only calling saveTask()
			// if the task title changed.
			if (this.task.title !== this.taskTitle) {
				this.$refs.taskTitle.blur()
				this.saveTask()
				this.taskTitle = taskTitle
			}
		},
		saveTask() {
			this.saving = true

			this.$store.dispatch('tasks/update', this.task)
				.then(() => {
					this.$emit('input', this.task)
					this.saved = true
					setTimeout(() => {
						this.saved = false
					}, 2000)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.saving = false
				})
		}
	},
}
</script>

