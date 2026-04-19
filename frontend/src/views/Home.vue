<script setup lang="ts">
import {ref, computed, watch, onMounted} from 'vue'

import {useDaytimeSalutation} from '@/composables/useDaytimeSalutation'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {useTaskStore} from '@/stores/tasks'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {formatDateSince, formatDisplayDate} from '@/helpers/time/formatDate'
import Message from '@/components/misc/Message.vue'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import TaskCollectionService from '@/services/taskCollection'
import TaskModel from '@/models/task'

const authStore = useAuthStore()
const projectStore = useProjectStore()
const taskStore = useTaskStore()

const salutation = useDaytimeSalutation()
const deletionScheduledAt = computed(() => parseDateOrNull(authStore.info?.deletionScheduledAt))

const username = computed(() => authStore.info?.name || authStore.info?.username || 'there')

const todayLabel = computed(() =>
	new Date().toLocaleDateString('en-US', {
		weekday: 'long',
		month: 'long',
		day: 'numeric',
	}),
)

const activeProjects = computed(() =>
	Object.values(projectStore.projects).filter(p => !p.isArchived),
)

// Task management
const newTaskTitle = ref('')
const displayedTasks = ref<ITask[]>([])
const showProjectModal = ref(false)
const selectedProject = ref<IProject | null>(null)
const isLoading = ref(false)

const pendingCount = computed(() => displayedTasks.value.filter(t => !t.done).length)
const doneCount = computed(() => displayedTasks.value.filter(t => t.done).length)

// Fetch tasks from backend
async function fetchTasks() {
	try {
		isLoading.value = true
		const taskCollectionService = new TaskCollectionService()
		// Fetch all tasks and filter to show recent/today's tasks
		const allTasks = await taskCollectionService.getAll(new TaskModel({}), {
			sort_by: ['due_date'],
			order_by: ['asc'],
			per_page: 50,
		})
		displayedTasks.value = allTasks.filter(t => !t.done)
	} catch (error) {
		console.error('Failed to fetch tasks:', error)
	} finally {
		isLoading.value = false
	}
}

// Show project selection modal
function initiateAddTask() {
	const title = newTaskTitle.value.trim()
	if (!title) return
	
	if (activeProjects.value.length === 0) {
		alert('Please create a project first')
		return
	}
	
	// If only one project, use it directly
	if (activeProjects.value.length === 1) {
		selectedProject.value = activeProjects.value[0]
		createTask()
	} else {
		// Show modal to select project
		showProjectModal.value = true
	}
}

// Create task with selected project
async function createTask() {
	if (!selectedProject.value) return
	
	const title = newTaskTitle.value.trim()
	if (!title) return
	
	try {
		isLoading.value = true
		await taskStore.createNewTask({
			title,
			projectId: selectedProject.value.id,
		})
		newTaskTitle.value = ''
		selectedProject.value = null
		showProjectModal.value = false
		await fetchTasks()
	} catch (error) {
		console.error('Failed to create task:', error)
	} finally {
		isLoading.value = false
	}
}

function selectProject(project: IProject) {
	selectedProject.value = project
	createTask()
}

function toggleTask(task: ITask) {
	task.done = !task.done
	taskStore.update(task)
	fetchTasks()
}

function priorityClass(priority?: number) {
	if (!priority) return 'chip-med'
	if (priority >= 3) return 'chip-high'
	if (priority === 2) return 'chip-med'
	return 'chip-low'
}

onMounted(() => {
	fetchTasks()
})

// Refresh tasks when showing modal closes
watch(() => showProjectModal.value, (isOpen) => {
	if (!isOpen) {
		fetchTasks()
	}
})
</script>

<template>
	<div class="vk-home">
		<!-- Deletion warning -->
		<Message
			v-if="deletionScheduledAt !== null"
			variant="danger"
			class="vk-deletion-msg"
		>
			{{
				$t('user.deletion.scheduled', {
					date: formatDisplayDate(deletionScheduledAt),
					dateSince: formatDateSince(deletionScheduledAt),
				})
			}}
			<RouterLink :to="{name: 'user.settings.deletion'}">
				{{ $t('user.deletion.scheduledCancel') }}
			</RouterLink>
		</Message>

		<!-- Topbar -->
		<header class="vk-topbar">
			<div>
				<h1 class="vk-topbar-title">
					{{ salutation }}, {{ username }}
				</h1>
				<p class="vk-topbar-sub">
					{{ todayLabel }} ·
					{{ pendingCount === 0 ? 'Nothing due today' : `${pendingCount} tasks pending` }}
				</p>
			</div>
			<div class="vk-topbar-actions">
				<RouterLink
					:to="{name: 'filters.index'}"
					class="vk-icon-btn"
					title="Filters"
					aria-label="Filters"
				>
					<svg
						width="14"
						height="14"
						viewBox="0 0 16 16"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
						stroke-linecap="round"
					>
						<path d="M2 4h12M4 8h8M6 12h4" />
					</svg>
				</RouterLink>
			</div>
		</header>

		<!-- Content -->
		<div class="vk-content">
			<!-- Import banner -->
			<div class="vk-import-banner">
				<span class="vk-import-star">✦</span>
				<p class="vk-import-text">
					<strong>Import your data</strong> — bring projects and tasks from Todoist, Trello, or Asana
				</p>
				<RouterLink
					:to="{name: 'migrate.start'}"
					class="vk-btn-ghost"
				>
					Import
				</RouterLink>
			</div>

			<!-- Quick add -->
			<div class="vk-quick-add">
				<input
					v-model="newTaskTitle"
					class="vk-input"
					type="text"
					placeholder="Add a task… (Enter to save)"
					@keydown.enter="initiateAddTask"
				>
				<button
					class="vk-btn-primary"
					:disabled="isLoading"
					@click="initiateAddTask"
				>
					{{ isLoading ? '⏳ Adding...' : '+ Add task' }}
				</button>
			</div>

			<!-- Stats -->
			<div class="vk-stats">
				<div class="vk-stat vk-stat--purple">
					<p class="vk-stat-num">
						{{ pendingCount }}
					</p>
					<p class="vk-stat-label">
						Tasks today
					</p>
				</div>
				<div class="vk-stat vk-stat--green">
					<p class="vk-stat-num">
						{{ doneCount }}
					</p>
					<p class="vk-stat-label">
						Completed this week
					</p>
				</div>
				<div class="vk-stat vk-stat--amber">
					<p class="vk-stat-num">
						{{ activeProjects.length }}
					</p>
					<p class="vk-stat-label">
						Projects active
					</p>
				</div>
			</div>

			<!-- Two-column body -->
			<div class="vk-two-col">
				<!-- Tasks panel -->
				<div class="vk-panel">
					<div class="vk-panel-head">
						<span class="vk-panel-title">Current Tasks</span>
						<RouterLink
							:to="{name: 'tasks.range'}"
							class="vk-panel-action"
						>
							View all
						</RouterLink>
					</div>

					<template v-if="displayedTasks.length === 0 && !isLoading">
						<div class="vk-empty">
							<div class="vk-empty-icon">
								🦙
							</div>
							<p class="vk-empty-title">
								Nothing to do — have a nice day!
							</p>
							<p class="vk-empty-sub">
								Your tasks will appear here once added
							</p>
						</div>
					</template>

					<template v-else-if="isLoading">
						<div class="vk-empty">
							<p class="vk-empty-title">
								Loading tasks...
							</p>
						</div>
					</template>

					<template v-else>
						<div
							v-for="task in displayedTasks"
							:key="task.id"
							class="vk-task-item"
							:class="{'vk-task-item--done': task.done}"
							@click="toggleTask(task)"
						>
							<div
								class="vk-check"
								:class="{'vk-check--done': task.done}"
							>
								<svg
									v-if="task.done"
									width="10"
									height="10"
									viewBox="0 0 10 10"
									fill="none"
									stroke="#fff"
									stroke-width="2"
									stroke-linecap="round"
								>
									<path d="M2 5l2.5 2.5L8 3" />
								</svg>
							</div>
							<span class="vk-task-text">{{ task.title }}</span>
							<span
								v-if="task.priority"
								class="vk-chip"
								:class="priorityClass(task.priority)"
							>
								{{ task.priority === 1 ? 'low' : task.priority === 2 ? 'med' : 'high' }}
							</span>
						</div>
					</template>
				</div>

				<!-- Right column -->
				<div class="vk-right-col">
					<!-- Projects -->
					<div class="vk-panel">
						<div class="vk-panel-head">
							<span class="vk-panel-title">Projects</span>
							<RouterLink
								:to="{name: 'project.create'}"
								class="vk-panel-action"
							>
								+ New
							</RouterLink>
						</div>
						<div
							v-if="activeProjects.length === 0"
							class="vk-empty"
							style="padding: 24px 20px"
						>
							<p class="vk-empty-title">
								No projects yet
							</p>
							<p class="vk-empty-sub">
								Create one to get started
							</p>
						</div>
						<div v-else>
							<RouterLink
								v-for="project in activeProjects"
								:key="project.id"
								:to="{name: 'project.index', params: {projectId: project.id}}"
								class="vk-project-item"
							>
								<span
									class="vk-project-dot"
									:style="{background: project.hexColor || '#6c63f5'}"
								/>
								<span class="vk-project-name">{{ project.title }}</span>
							</RouterLink>
						</div>
					</div>

					<!-- Upcoming -->
					<div class="vk-panel">
						<div class="vk-panel-head">
							<span class="vk-panel-title">Upcoming</span>
							<RouterLink
								:to="{name: 'tasks.range'}"
								class="vk-panel-action"
							>
								View all
							</RouterLink>
						</div>
						<div
							class="vk-empty"
							style="padding: 24px 20px"
						>
							<p class="vk-empty-title">
								No upcoming deadlines
							</p>
							<p class="vk-empty-sub">
								Enjoy the calm
							</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Project Selection Modal -->
			<div
				v-if="showProjectModal"
				class="vk-modal-overlay"
				@click.self="showProjectModal = false"
			>
				<div class="vk-modal">
					<div class="vk-modal-header">
						<h3>Select Project</h3>
						<button
							class="vk-modal-close"
							@click="showProjectModal = false"
						>
							✕
						</button>
					</div>
					<div class="vk-modal-content">
						<p class="vk-modal-subtitle">
							Choose which project to add "<strong>{{ newTaskTitle }}</strong>" to:
						</p>
						<div class="vk-project-list">
							<button
								v-for="project in activeProjects"
								:key="project.id"
								class="vk-project-option"
								@click="selectProject(project)"
							>
								<span
									class="vk-project-dot"
									:style="{background: project.hexColor || '#6c63f5'}"
								/>
								<span class="vk-project-name">{{ project.title }}</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@300;400;500;600&family=Playfair+Display:wght@400;500&display=swap');

.vk-home *,
.vk-home *::before,
.vk-home *::after {
	box-sizing: border-box;
}

.vk-home {
	--bg: #0e0f13;
	--bg-panel: #13141a;
	--bg-hover: #181920;
	--bg-active: #1e1b3a;
	--border: #1e1f28;
	--border-mid: #2a2b35;
	--text-primary: #f0ede8;
	--text-secondary: #8a8a9a;
	--text-muted: #4a4b57;
	--accent: #6c63f5;
	--accent-light: #a78bfa;
	--green: #10b981;
	--amber: #f59e0b;

	font-family: 'DM Sans', sans-serif;
	background: var(--bg);
	color: var(--text-primary);
	display: flex;
	flex-direction: column;
	min-height: 100%;
	font-size: 16px;
	line-height: 1.5;
}

.vk-deletion-msg {
	margin: 16px 32px 0;
}

/* ─── Topbar ─── */
.vk-topbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 18px 32px;
	padding-top: calc(4rem + 18px); /* clear the fixed navbar */
	border-bottom: 0.5px solid var(--border);
	background: var(--bg);
}

.vk-topbar-title {
	font-family: 'Playfair Display', serif;
	font-size: 26px;
	font-weight: 400;
	color: var(--text-primary);
	letter-spacing: 0.01em;
	line-height: 1.2;
	border: none;
	margin: 0;
	padding: 0;
}

.vk-topbar-sub {
	font-size: 14px;
	color: var(--text-muted);
	margin-top: 3px;
}

.vk-topbar-actions {
	display: flex;
	align-items: center;
	gap: 8px;
}

.vk-icon-btn {
	width: 34px;
	height: 34px;
	border-radius: 8px;
	background: var(--bg-panel);
	border: 0.5px solid var(--border-mid);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	color: var(--text-muted);
	text-decoration: none;
	transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.vk-icon-btn:hover {
	background: #22232e;
	color: var(--accent-light);
	border-color: var(--accent);
}

/* ─── Content ─── */
.vk-content {
	flex: 1;
	padding: 28px 32px;
	overflow-y: auto;
}

.vk-import-banner {
	display: flex;
	align-items: center;
	gap: 12px;
	background: #1a1b3a;
	border: 0.5px solid #2e2b55;
	border-radius: 10px;
	padding: 12px 16px;
	margin-bottom: 22px;
}

.vk-import-star {
	font-size: 18px;
	color: var(--accent-light);
	flex-shrink: 0;
}

.vk-import-text {
	font-size: 14px;
	color: #7070a0;
	flex: 1;
}

.vk-import-text strong {
	color: #a0a0c0;
	font-weight: 500;
}

.vk-quick-add {
	display: flex;
	gap: 10px;
	margin-bottom: 24px;
}

.vk-input {
	flex: 1;
	background: var(--bg-panel);
	border: 0.5px solid var(--border-mid);
	border-radius: 10px;
	padding: 12px 16px;
	font-size: 15px;
	color: #d0cdc8;
	font-family: 'DM Sans', sans-serif;
	outline: none;
	transition: border-color 0.15s;
	box-shadow: none;
	height: auto;
}

.vk-input::placeholder {
	color: #3a3b48;
}

.vk-input:focus {
	border-color: var(--accent);
	box-shadow: none;
}

.vk-btn-primary {
	background: linear-gradient(135deg, var(--accent), #8b83f7);
	color: #fff;
	border: none;
	border-radius: 10px;
	padding: 0 20px;
	font-size: 15px;
	font-weight: 500;
	cursor: pointer;
	font-family: 'DM Sans', sans-serif;
	letter-spacing: 0.03em;
	transition: opacity 0.15s;
	white-space: nowrap;
	height: 44px;
}

.vk-btn-primary:hover {
	opacity: 0.85;

.vk-btn-primary:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}
}

.vk-btn-ghost {
	background: transparent;
	border: 0.5px solid #4040a0;
	color: var(--accent-light);
	border-radius: 6px;
	padding: 6px 14px;
	font-size: 13px;
	cursor: pointer;
	font-family: 'DM Sans', sans-serif;
	white-space: nowrap;
	text-decoration: none;
	transition: background 0.15s;
}

.vk-btn-ghost:hover {
	background: var(--bg-active);
}

/* ─── Stats ─── */
.vk-stats {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	gap: 12px;
	margin-bottom: 24px;
}

.vk-stat {
	background: var(--bg-panel);
	border: 0.5px solid var(--border);
	border-radius: 12px;
	padding: 16px 20px;
	position: relative;
	overflow: hidden;
}

.vk-stat::before {
	content: '';
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	height: 2px;
}

.vk-stat--purple::before { background: linear-gradient(90deg, var(--accent), transparent); }
.vk-stat--green::before  { background: linear-gradient(90deg, var(--green), transparent); }
.vk-stat--amber::before  { background: linear-gradient(90deg, var(--amber), transparent); }

.vk-stat-num {
	font-size: 32px;
	font-weight: 300;
	color: var(--text-primary);
	line-height: 1;
	margin-bottom: 4px;
	font-family: 'Playfair Display', serif;
}

.vk-stat-label {
	font-size: 13px;
	color: var(--text-muted);
}

/* ─── Two-column layout ─── */
.vk-two-col {
	display: grid;
	grid-template-columns: 1fr 300px;
	gap: 20px;
}

.vk-right-col {
	display: flex;
	flex-direction: column;
	gap: 16px;
}

/* ─── Panel ─── */
.vk-panel {
	background: var(--bg-panel);
	border: 0.5px solid var(--border);
	border-radius: 12px;
	overflow: hidden;
}

.vk-panel-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 14px 18px;
	border-bottom: 0.5px solid var(--border);
}

.vk-panel-title {
	font-size: 15px;
	font-weight: 500;
	color: #c0bdb8;
}

.vk-panel-action {
	font-size: 13px;
	color: var(--accent);
	cursor: pointer;
	text-decoration: none;
	transition: opacity 0.15s;
}

.vk-panel-action:hover {
	opacity: 0.75;
}

/* ─── Empty state ─── */
.vk-empty {
	padding: 48px 20px;
	text-align: center;
}

.vk-empty-icon {
	width: 64px;
	height: 64px;
	margin: 0 auto 14px;
	background: var(--bg-active);
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 28px;
}

.vk-empty-title {
	font-size: 15px;
	color: var(--text-secondary);
	margin-bottom: 4px;
}

/* ─── Modal ─── */
.vk-modal-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.7);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	animation: fadeIn 0.2s ease-in-out;
}

@keyframes fadeIn {
	from { opacity: 0; }
	to { opacity: 1; }
}

.vk-modal {
	background: var(--bg-panel);
	border: 0.5px solid var(--border);
	border-radius: 12px;
	max-width: 400px;
	width: 90%;
	max-height: 80vh;
	overflow-y: auto;
	box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
	animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
	from {
		transform: translateY(20px);
		opacity: 0;
	}
	to {
		transform: translateY(0);
		opacity: 1;
	}
}

.vk-modal-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20px 24px;
	border-bottom: 0.5px solid var(--border);
}

.vk-modal-header h3 {
	margin: 0;
	font-size: 18px;
	color: var(--text-primary);
	font-weight: 500;
}

.vk-modal-close {
	background: transparent;
	border: none;
	color: var(--text-muted);
	font-size: 20px;
	cursor: pointer;
	padding: 0;
	width: 32px;
	height: 32px;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: color 0.15s;
}

.vk-modal-close:hover {
	color: var(--text-primary);
}

.vk-modal-content {
	padding: 24px;
}

.vk-modal-subtitle {
	margin: 0 0 16px 0;
	font-size: 14px;
	color: var(--text-muted);
}

.vk-modal-subtitle strong {
	color: var(--text-primary);
}

.vk-project-list {
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.vk-project-option {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 12px 14px;
	background: var(--bg-active);
	border: 0.5px solid var(--border-mid);
	border-radius: 8px;
	cursor: pointer;
	color: var(--text-secondary);
	transition: all 0.15s;
	text-align: left;
	font-family: 'DM Sans', sans-serif;
	font-size: 15px;
}

.vk-project-option:hover {
	background: var(--bg-hover);
	border-color: var(--accent);
	color: var(--text-primary);
}

.vk-project-option .vk-project-dot {
	width: 10px;
	height: 10px;
}

.vk-project-option .vk-project-name {
	flex: 1;
	color: inherit;
}

.vk-empty-sub {
	font-size: 13px;
	color: #3a3b48;
}

/* ─── Task items ─── */
.vk-task-item {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 12px 18px;
	border-bottom: 0.5px solid #1a1b22;
	cursor: pointer;
	transition: background 0.1s;
}

.vk-task-item:hover { background: var(--bg-hover); }
.vk-task-item:last-child { border-bottom: none; }

.vk-task-item--done .vk-task-text {
	text-decoration: line-through;
	opacity: 0.4;
}

.vk-check {
	width: 17px;
	height: 17px;
	border-radius: 50%;
	border: 1.5px solid #3a3b48;
	flex-shrink: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.15s;
}

.vk-task-item:hover .vk-check { border-color: var(--accent); }

.vk-check--done {
	background: var(--accent);
	border-color: var(--accent);
}

.vk-task-text {
	font-size: 15px;
	color: #a0a0b0;
	flex: 1;
}

.vk-chip {
	font-size: 12px;
	padding: 2px 8px;
	border-radius: 20px;
	border: 0.5px solid;
	text-transform: capitalize;
}

.chip-high  { border-color: rgba(224,72,72,.3);   color: #f87171; background: rgba(224,72,72,.08); }
.chip-med   { border-color: rgba(108,99,245,.3);  color: var(--accent-light); background: rgba(108,99,245,.08); }
.chip-low   { border-color: rgba(16,185,129,.3);  color: #34d399; background: rgba(16,185,129,.08); }

/* ─── Projects list ─── */
.vk-project-item {
	display: flex;
	align-items: center;
	gap: 10px;
	padding: 10px 16px;
	cursor: pointer;
	text-decoration: none;
	transition: background 0.1s;
}

.vk-project-item:hover { background: var(--bg-hover); }

.vk-project-dot {
	width: 8px;
	height: 8px;
	border-radius: 50%;
	flex-shrink: 0;
}

.vk-project-name {
	font-size: 15px;
	color: var(--text-secondary);
	flex: 1;
}
</style>
