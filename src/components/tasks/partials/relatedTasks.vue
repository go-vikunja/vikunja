<template>
	<div class="task-relations">
		<x-button
			v-if="Object.keys(relatedTasks).length > 0"
			@click="showNewRelationForm = !showNewRelationForm"
			class="is-pulled-right add-task-relation-button"
			:class="{'is-active': showNewRelationForm}"
			v-tooltip="'Add a New Task Relation'"
			type="secondary"
			icon="plus"
		/>
		<transition-group name="fade">
			<template v-if="editEnabled && showCreate">
				<label class="label" key="label">
					New Task Relation
					<transition name="fade">
						<span class="is-inline-flex" v-if="taskRelationService.loading">
							<span class="loader is-inline-block mr-2"></span>
							Saving...
						</span>
						<span class="has-text-success" v-if="!taskRelationService.loading && saved">
							Saved!
						</span>
					</transition>
				</label>
				<div class="field" key="field-search">
					<multiselect
						placeholder="Type search for a new task to add as related..."
						@search="findTasks"
						:loading="taskService.loading"
						:search-results="foundTasks"
						label="title"
						v-model="newTaskRelationTask"
						:creatable="true"
						create-placeholder="Add this as new related task"
						@create="createAndRelateTask"
					/>
				</div>
				<div class="field has-addons mb-4" key="field-kind">
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
						<x-button @click="addTaskRelation()">Add Task Relation</x-button>
					</div>
				</div>
			</template>
		</transition-group>

		<div :key="kind" class="related-tasks" v-for="(rts, kind ) in relatedTasks">
			<template v-if="rts.length > 0">
				<span class="title">{{ relationKindTitle(kind, rts.length) }}</span>
				<div class="tasks noborder">
					<div :key="t.id" class="task" v-for="t in rts.filter(t => t)">
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

import Multiselect from '@/components/input/multiselect'

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
			saved: false,
			showNewRelationForm: false,
		}
	},
	components: {
		Multiselect,
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
	computed: {
		showCreate() {
			return Object.keys(this.relatedTasks).length === 0 || this.showNewRelationForm
		},
	},
	methods: {
		findTasks(query) {
			this.taskService.getAll({}, {s: query})
				.then(response => {
					this.$set(this, 'foundTasks', response)
				})
				.catch(e => {
					this.error(e, this)
				})
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
					this.newTaskRelationTask = null
					this.saved = true
					this.showNewRelationForm = false
					setTimeout(() => {
						this.saved = false
					}, 2000)
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
					this.saved = true
					setTimeout(() => {
						this.saved = false
					}, 2000)
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

<style lang="scss">
@import '@/styles/theme/variables';

.add-task-relation-button {
	margin-top: -3rem;

	svg {
		transition: transform $transition;
	}

	&.is-active svg {
		transform: rotate(45deg);
	}
}
</style>
