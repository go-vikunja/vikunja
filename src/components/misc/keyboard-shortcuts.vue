<template>
	<div class="modal-mask keyboard-shortcuts-modal">
		<div @click.self="close()" class="modal-container">
			<div class="modal-content">
				<div class="card has-background-white has-no-shadow">
					<header class="card-header">
						<p class="card-header-title">Available Keyboard Shortcuts</p>
					</header>
					<div class="card-content content">
						<p>
							<strong>Toggle The Menu</strong>
							<span class="shortcuts">
								<span>ctrl</span>
								<i>+</i>
								<span>e</span>
							</span>
						</p>
						<h3>Kanban</h3>
						<div class="message is-primary" v-if="$route.name === 'list.kanban'">
							<div class="message-body">
								These shortcuts work on the current page.
							</div>
						</div>
						<p>
							<strong>Mark a task as done</strong>
							<span class="shortcuts">
								<span>ctrl</span>
								<i>+</i>
								<span>click</span>
							</span>
						</p>
						<h3>Task Page</h3>
						<div class="message is-primary" v-if="$route.name === 'task.detail' || $route.name === 'task.list.detail' || $route.name === 'task.gantt.detail' || $route.name === 'task.kanban.detail' || $route.name === 'task.detail'">
							<div class="message-body">
								These shortcuts work on the current page.
							</div>
						</div>
						<p>
							<strong>Assign this task to a user</strong>
							<span class="shortcuts">
								<span>a</span>
							</span>
						</p>
						<p>
							<strong>Add labels to this task</strong>
							<span class="shortcuts">
								<span>l</span>
							</span>
						</p>
						<p>
							<strong>Change the due date of this task</strong>
							<span class="shortcuts">
								<span>d</span>
							</span>
						</p>
						<p>
							<strong>Add an attachment to this task</strong>
							<span class="shortcuts">
								<span>f</span>
							</span>
						</p>
						<p>
							<strong>Modify related tasks of this task</strong>
							<span class="shortcuts">
								<span>r</span>
							</span>
						</p>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import {KEYBOARD_SHORTCUTS_ACTIVE} from '@/store/mutation-types'

export default {
	name: 'keyboard-shortcuts',
	methods: {
		close() {
			this.$store.commit(KEYBOARD_SHORTCUTS_ACTIVE, false)
		},
	},
}
</script>
