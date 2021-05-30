<template>
	<div class="modal-mask keyboard-shortcuts-modal">
		<div @click.self="close()" class="modal-container">
			<div class="modal-content">
				<card class="has-background-white has-no-shadow" title="Available Keyboard Shortcuts">
					<div class="message is-primary">
						<div class="message-body">
							These shortcuts work on all pages.
						</div>
					</div>
					<p>
						<strong>Toggle The Menu</strong>
						<shortcut :keys="['ctrl', 'e']"/>
					</p>
					<p>
						<strong>Open the search/quick action bar</strong>
						<shortcut :keys="['ctrl', 'k']"/>
					</p>
					<h3>Kanban</h3>
					<div class="message is-primary" v-if="$route.name === 'list.kanban'">
						<div class="message-body">
							These shortcuts work only on the current page.
						</div>
					</div>
					<p>
						<strong>Mark a task as done</strong>
						<shortcut :keys="['ctrl', 'click']"/>
					</p>
					<h3>Task Page</h3>
					<div
						class="message is-primary"
						v-if="$route.name === 'task.detail' || $route.name === 'task.list.detail' || $route.name === 'task.gantt.detail' || $route.name === 'task.kanban.detail' || $route.name === 'task.detail'">
						<div class="message-body">
							These shortcuts work only on the current page.
						</div>
					</div>
					<p>
						<strong>Assign this task to a user</strong>
						<shortcut :keys="['a']"/>
					</p>
					<p>
						<strong>Add labels to this task</strong>
						<shortcut :keys="['l']"/>
					</p>
					<p>
						<strong>Change the due date of this task</strong>
						<shortcut :keys="['d']"/>
					</p>
					<p>
						<strong>Add an attachment to this task</strong>
						<shortcut :keys="['f']"/>
					</p>
					<p>
						<strong>Modify related tasks of this task</strong>
						<shortcut :keys="['r']"/>
					</p>
				</card>
			</div>
		</div>
	</div>
</template>

<script>
import {KEYBOARD_SHORTCUTS_ACTIVE} from '@/store/mutation-types'
import Shortcut from '@/components/misc/shortcut'

export default {
	name: 'keyboard-shortcuts',
	components: {Shortcut},
	methods: {
		close() {
			this.$store.commit(KEYBOARD_SHORTCUTS_ACTIVE, false)
		},
	},
}
</script>
