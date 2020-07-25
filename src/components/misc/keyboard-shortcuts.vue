<template>
	<div class="modal-mask keyboard-shortcuts-modal">
		<div class="modal-container" @click.self="close()">
			<div class="modal-content">
				<div class="card has-background-white has-no-shadow">
					<header class="card-header">
						<p class="card-header-title">Available Keyboard Shortcuts</p>
					</header>
					<div class="card-content content">
						<p class="info">
							The available keyboard shortcuts depend on the current page. Not all shortcuts are available
							everywhere.
						</p>
						<p>
							<strong>Toggle The Menu</strong>
							<span class="shortcuts">
								<span>ctrl</span>
								<i>+</i>
								<span>e</span>
							</span>
						</p>
						<p v-if="$route.name === 'list.kanban'">
							<strong>Mark a task as done</strong>
							<span class="shortcuts">
								<span>ctrl</span>
								<i>+</i>
								<span>click</span>
							</span>
						</p>
						<template
								v-if="$route.name === 'task.detail' || $route.name === 'task.list.detail' || $route.name === 'task.gantt.detail' || $route.name === 'task.kanban.detail' || $route.name === 'task.detail'">
							<p>
								<strong>Assign this task to a user</strong>
								<span class="shortcuts">
									<span>ctrl</span>
									<i>+</i>
									<span>shift</span>
									<i>+</i>
									<span>a</span>
								</span>
							</p>
							<p>
								<strong>Add labels to this task</strong>
								<span class="shortcuts">
									<span>ctrl</span>
									<i>+</i>
									<span>shift</span>
									<i>+</i>
									<span>l</span>
								</span>
							</p>
							<p>
								<strong>Change the due date of this task</strong>
								<span class="shortcuts">
									<span>ctrl</span>
									<i>+</i>
									<span>shift</span>
									<i>+</i>
									<span>d</span>
								</span>
							</p>
							<p>
								<strong>Add an attachment to this task</strong>
								<span class="shortcuts">
									<span>ctrl</span>
									<i>+</i>
									<span>shift</span>
									<i>+</i>
									<span>f</span>
								</span>
							</p>
							<p>
								<strong>Modify related tasks of this task</strong>
								<span class="shortcuts">
									<span>ctrl</span>
									<i>+</i>
									<span>shift</span>
									<i>+</i>
									<span>r</span>
								</span>
							</p>
						</template>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	export default {
		name: 'keyboard-shortcuts',
		methods: {
			close() {
				this.$emit('close')
			},
		},
	}
</script>
