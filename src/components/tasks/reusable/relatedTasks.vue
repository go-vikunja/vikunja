<template>
	<div class="task-relations">
		<label class="label">New Task Relation</label>
		<div class="columns">
			<div class="column is-three-quarters">
				<multiselect
						v-model="newTaskRelationTask"
						:options="foundTasks"
						:multiple="false"
						:searchable="true"
						:loading="taskService.loading"
						:internal-search="true"
						@search-change="findTasks"
						placeholder="Type search for a new task to add as related..."
						label="text"
						track-by="id"
						:taggable="true"
						:showNoOptions="false"
						@tag="createAndRelateTask"
						tag-placeholder="Add this as new related task"
				>
					<template slot="clear" slot-scope="props">
						<div class="multiselect__clear"
							v-if="newTaskRelationTask !== null && newTaskRelationTask.id !== 0"
							@mousedown.prevent.stop="clearAllFoundTasks(props.search)"></div>
					</template>
					<span slot="noResult">No task found. Consider changing the search query.</span>
				</multiselect>
			</div>
			<div class="column field has-addons">
				<div class="control is-expanded">
					<div class="select is-fullwidth has-defaults">
						<select v-model="newTaskRelationKind">
							<option value="unset">Select a relation kind</option>
							<option v-for="(label, rk) in relationKinds" :key="rk" :value="rk">
								{{ label }}
							</option>
						</select>
					</div>
				</div>
				<div class="control">
					<a class="button is-primary" @click="addTaskRelation()">Add task Relation</a>
				</div>
			</div>
		</div>

		<div class="related-tasks" v-for="(rts, kind ) in relatedTasks" :key="kind">
			<template v-if="rts.length > 0">
				<span class="title">{{ relationKindTitle(kind, rts.length) }}</span>
				<div class="tasks noborder">
					<div class="task" v-for="t in rts" :key="t.id">
						<router-link :to="{ name: 'taskDetailView', params: { id: t.id } }">
							<span class="tasktext" :class="{ 'done': t.done}">
								{{t.text}}
							</span>
						</router-link>
						<a class="remove" @click="() => {showDeleteModal = true; relationToDelete = {relationKind: kind, otherTaskId: t.id}}">
							<icon icon="trash-alt"/>
						</a>
					</div>
				</div>
			</template>
		</div>
		<p v-if="showNoRelationsNotice && Object.keys(relatedTasks).length === 0" class="none">
			No task relations yet.
		</p>

		<!-- Delete modal -->
		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="removeTaskRelation()">
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

	import multiselect from 'vue-multiselect'

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
			multiselect,
		},
		props: {
			taskID: {
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
			}
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
					taskId: this.taskID,
					otherTaskId: this.newTaskRelationTask.id,
					relationKind: this.newTaskRelationKind,
				})
				this.taskRelationService.create(rel)
					.then(() => {
						if (!this.relatedTasks[this.newTaskRelationKind]) {
							this.$set(this.relatedTasks, this.newTaskRelationKind, [])
						}
						this.relatedTasks[this.newTaskRelationKind].push(this.newTaskRelationTask)
						this.newTaskRelationKind = 'unset'
						this.newTaskRelationTask = new TaskModel()
						this.success({message: 'The task relation was created successfully'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			removeTaskRelation() {
				let rel = new TaskRelationModel({
					relationKind: this.relationToDelete.relationKind,
					taskId: this.taskID,
					otherTaskId: this.relationToDelete.otherTaskId,
				})
				this.taskRelationService.delete(rel)
					.then(r => {
						Object.keys(this.relatedTasks).forEach(relationKind => {
							for (const t in this.relatedTasks[relationKind]) {
								if (this.relatedTasks[relationKind][t].id === this.relationToDelete.otherTaskId && relationKind === this.relationToDelete.relationKind) {
									this.relatedTasks[relationKind].splice(t, 1)
								}
							}
						})
						this.success(r, this)
					})
					.catch(e => {
						this.error(e, this)
					})
					.finally(() => {
						this.showDeleteModal = false
					})
			},
			createAndRelateTask(text) {
				const newTask = new TaskModel({text: text, listId: this.listId})
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
			}
		},
	}
</script>
