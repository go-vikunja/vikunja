<template>
	<div class="task-relations">
		<template v-if="editEnabled">
			<label class="label">New Task Relation</label>
			<div class="field">
				<multiselect
					:internal-search="true"
					:loading="taskService.loading"
					:multiple="false"
					:options="foundTasks"
					:searchable="true"
					:showNoOptions="false"
					:taggable="true"
					@search-change="findTasks"
					@tag="createAndRelateTask"
					label="title"
					placeholder="Type search for a new task to add as related..."
					tag-placeholder="Add this as new related task"
					track-by="id"
					v-model="newTaskRelationTask"
				>
					<template slot="clear" slot-scope="props">
						<div
							@mousedown.prevent.stop="clearAllFoundTasks(props.search)"
							class="multiselect__clear"
							v-if="newTaskRelationTask !== null && newTaskRelationTask.id !== 0"></div>
					</template>
					<span slot="noResult">No task found. Consider changing the search query.</span>
				</multiselect>
			</div>
			<div class="field has-addons">
				<div class="control is-expanded">
					<div class="select is-fullwidth has-defaults">
						<select v-model="newTaskRelationKind">
							<option value="unset">Select a relation kind</option>
							<option :key="rk" :value="rk" v-for="(label, rk) in relationKinds">
								{{ label[0] }}
							</option>
						</select>
					</div>
				</div>
				<div class="control">
					<a @click="addTaskRelation()" class="button is-primary">Add task Relation</a>
				</div>
			</div>
		</template>

		<div :key="kind" class="related-tasks" v-for="(rts, kind ) in relatedTasks">
			<template v-if="rts.length > 0">
				<span class="title">{{ relationKindTitle(kind, rts.length) }}</span>
				<div class="tasks noborder">
					<div :key="t.id" class="task" v-for="t in rts">
						<router-link :to="{ name: $route.name, params: { id: t.id } }">
							<span :class="{ 'done': t.done}" class="tasktext">
								<span
									class="different-list"
									v-if="t.listId !== listId"
									v-tooltip="'This task belongs to a different list.'">
									{{
										$store.getters['lists/getListById'](t.listId) === null ? '' : $store.getters['lists/getListById'](t.listId).title
									}} >
								</span>
								{{ t.title }}
							</span>
						</router-link>
						<a
							@click="() => {showDeleteModal = true; relationToDelete = {relationKind: kind, otherTaskId: t.id}}"
							class="remove"
							v-if="editEnabled">
							<icon icon="trash-alt"/>
						</a>
					</div>
				</div>
			</template>
		</div>
		<p class="none" v-if="showNoRelationsNotice && Object.keys(relatedTasks).length === 0">
			No task relations yet.
		</p>

		<!-- Delete modal -->
		<modal
			@close="showDeleteModal = false"
			@submit="removeTaskRelation()"
			v-if="showDeleteModal">
			<span slot="header">Delete Task Relation</span>
			<p slot="text">Are you sure you want to delete this task relation?<br/>
				<b>This CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
import TaskService from '../../../services/task'
import TaskModel from '../../../models/task'
import TaskRelationService from '../../../services/taskRelation'
import relationKinds from '../../../models/relationKinds'
import TaskRelationModel from '../../../models/taskRelation'

import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'

export default {
	name: 'relatedTasks',
	data() {
		return {
			relatedTasks: {},
			taskService: TaskService,
			foundTasks: [],
			relationKinds: relationKinds,
			newTaskRelationTask: TaskModel,
			newTaskRelationKind: 'related',
			taskRelationService: TaskRelationService,
			showDeleteModal: false,
			relationToDelete: {},
		}
	},
	components: {
		multiselect: () => ({
			component: import(/* webpackPrefetch: true *//* webpackChunkName: "multiselect" */ 'vue-multiselect'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	props: {
		taskId: {
			type: Number,
			required: true,
		},
		initialRelatedTasks: {
			type: Object,
			default: () => {
			},
		},
		showNoRelationsNotice: {
			type: Boolean,
			default: false,
		},
		listId: {
			type: Number,
			default: 0,
		},
		editEnabled: {
			default: true,
		},
	},
	created() {
		this.taskService = new TaskService()
		this.taskRelationService = new TaskRelationService()
		this.newTaskRelationTask = new TaskModel()
	},
	watch: {
		initialRelatedTasks(newVal) {
			this.relatedTasks = newVal
		},
	},
	mounted() {
		this.relatedTasks = this.initialRelatedTasks
	},
	methods: {
		findTasks(query) {
			if (query === '') {
				this.clearAllFoundTasks()
				return
			}

			this.taskService.getAll({}, {s: query})
				.then(response => {
					this.$set(this, 'foundTasks', response)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		clearAllFoundTasks() {
			this.$set(this, 'foundTasks', [])
		},
		addTaskRelation() {
			let rel = new TaskRelationModel({
				taskId: this.taskId,
				otherTaskId: this.newTaskRelationTask.id,
				relationKind: this.newTaskRelationKind,
			})
			this.taskRelationService.create(rel)
				.then(() => {
					if (!this.relatedTasks[this.newTaskRelationKind]) {
						this.$set(this.relatedTasks, this.newTaskRelationKind, [])
					}
					this.relatedTasks[this.newTaskRelationKind].push(this.newTaskRelationTask)
					this.newTaskRelationTask = new TaskModel()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		removeTaskRelation() {
			let rel = new TaskRelationModel({
				relationKind: this.relationToDelete.relationKind,
				taskId: this.taskId,
				otherTaskId: this.relationToDelete.otherTaskId,
			})
			this.taskRelationService.delete(rel)
				.then(() => {
					Object.keys(this.relatedTasks).forEach(relationKind => {
						for (const t in this.relatedTasks[relationKind]) {
							if (this.relatedTasks[relationKind][t].id === this.relationToDelete.otherTaskId && relationKind === this.relationToDelete.relationKind) {
								this.relatedTasks[relationKind].splice(t, 1)
							}
						}
					})
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.showDeleteModal = false
				})
		},
		createAndRelateTask(title) {
			const newTask = new TaskModel({title: title, listId: this.listId})
			this.taskService.create(newTask)
				.then(r => {
					this.newTaskRelationTask = r
					this.addTaskRelation()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		relationKindTitle(kind, length) {
			if (length > 1) {
				return relationKinds[kind][1]
			}
			return relationKinds[kind][0]
		},
	},
}
</script>
