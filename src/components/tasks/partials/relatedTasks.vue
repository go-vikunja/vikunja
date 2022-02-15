<template>
	<div class="task-relations">
		<x-button
			v-if="editEnabled && Object.keys(relatedTasks).length > 0"
			@click="showNewRelationForm = !showNewRelationForm"
			class="is-pulled-right add-task-relation-button"
			:class="{'is-active': showNewRelationForm}"
			v-tooltip="$t('task.relation.add')"
			variant="secondary"
			icon="plus"
			:shadow="false"
		/>
		<transition-group name="fade">
			<template v-if="editEnabled && showCreate">
				<label class="label" key="label">
					{{ $t('task.relation.new') }}
					<transition name="fade">
						<span class="is-inline-flex" v-if="taskRelationService.loading">
							<span class="loader is-inline-block mr-2"></span>
							{{ $t('misc.saving') }}
						</span>
						<span class="has-text-success" v-else-if="!taskRelationService.loading && saved">
							{{ $t('misc.saved') }}
						</span>
					</transition>
				</label>
				<div class="field" key="field-search">
					<multiselect
						:placeholder="$t('task.relation.searchPlaceholder')"
						@search="findTasks"
						:loading="taskService.loading"
						:search-results="mappedFoundTasks"
						label="title"
						v-model="newTaskRelationTask"
						:creatable="true"
						:create-placeholder="$t('task.relation.createPlaceholder')"
						@create="createAndRelateTask"
					>
						<template #searchResult="props">
							<span v-if="typeof props.option !== 'string'" class="search-result">
								<span
									class="different-list"
									v-if="props.option.listId !== listId"
								>
									<span
										v-if="props.option.differentNamespace !== null"
										v-tooltip="$t('task.relation.differentNamespace')">
										{{ props.option.differentNamespace }} >
									</span>
									<span
										v-if="props.option.differentList !== null"
										v-tooltip="$t('task.relation.differentList')">
										{{ props.option.differentList }} >
									</span>
								</span>
								{{ props.option.title }}
							</span>
							<span class="search-result" v-else>
								{{ props.option }}
							</span>
						</template>
					</multiselect>
				</div>
				<div class="field has-addons mb-4" key="field-kind">
					<div class="control is-expanded">
						<div class="select is-fullwidth has-defaults">
							<select v-model="newTaskRelationKind">
								<option value="unset">{{ $t('task.relation.select') }}</option>
								<option :key="rk" :value="rk" v-for="rk in relationKinds">
									{{ $tc(`task.relation.kinds.${rk}`, 1) }}
								</option>
							</select>
						</div>
					</div>
					<div class="control">
						<x-button @click="addTaskRelation()">{{ $t('task.relation.add') }}</x-button>
					</div>
				</div>
			</template>
		</transition-group>

		<div :key="rts.kind" class="related-tasks" v-for="rts in mappedRelatedTasks">
			<span class="title">{{ rts.title }}</span>
			<div class="tasks">
				<div :key="t.id" class="task" v-for="t in rts.tasks">
					<router-link
						:to="{ name: $route.name, params: { id: t.id } }"
						:class="{ 'is-strikethrough': t.done}">
						<span
							class="different-list"
							v-if="t.listId !== listId"
						>
							<span
								v-if="t.differentNamespace !== null"
								v-tooltip="$t('task.relation.differentNamespace')">
								{{ t.differentNamespace }} >
							</span>
							<span
								v-if="t.differentList !== null"
								v-tooltip="$t('task.relation.differentList')">
								{{ t.differentList }} >
							</span>
						</span>
						{{ t.title }}
					</router-link>
					<a
						@click="() => {showDeleteModal = true; relationToDelete = {relationKind: rts.kind, otherTaskId: t.id}}"
						class="remove"
						v-if="editEnabled">
						<icon icon="trash-alt"/>
					</a>
				</div>
			</div>
		</div>
		<p class="none" v-if="showNoRelationsNotice && Object.keys(relatedTasks).length === 0">
			{{ $t('task.relation.noneYet') }}
		</p>

		<!-- Delete modal -->
		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				@submit="removeTaskRelation()"
				v-if="showDeleteModal"
			>
				<template #header><span>{{ $t('task.relation.delete') }}</span></template>

				<template #text>
					<p>{{ $t('task.relation.deleteText1') }}<br/>
						<strong>{{ $t('task.relation.deleteText2') }}</strong></p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import TaskService from '../../../services/task'
import TaskModel from '../../../models/task'
import TaskRelationService from '../../../services/taskRelation'
import relationKinds from '../../../models/constants/relationKinds'
import TaskRelationModel from '../../../models/taskRelation'

import Multiselect from '@/components/input/multiselect.vue'

export default defineComponent({
	name: 'relatedTasks',
	data() {
		return {
			relatedTasks: {},
			taskService: new TaskService(),
			foundTasks: [],
			relationKinds: relationKinds,
			newTaskRelationTask: new TaskModel(),
			newTaskRelationKind: 'related',
			taskRelationService: new TaskRelationService(),
			showDeleteModal: false,
			relationToDelete: {},
			saved: false,
			showNewRelationForm: false,
			query: '',
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
	watch: {
		initialRelatedTasks: {
			handler(value) {
				this.relatedTasks = value
			},
			immediate: true,
		},
	},
	computed: {
		showCreate() {
			return Object.keys(this.relatedTasks).length === 0 || this.showNewRelationForm
		},
		namespace() {
			return this.$store.getters['namespaces/getListAndNamespaceById'](this.listId, true)?.namespace
		},
		mappedRelatedTasks() {
			return Object.entries(this.relatedTasks).map(([kind, tasks]) => ({
				title: this.$tc(`task.relation.kinds.${kind}`, tasks.length),
				tasks: this.mapRelatedTasks(tasks),
				kind,
			}))
		},
		mappedFoundTasks() {
			return this.mapRelatedTasks(this.foundTasks.filter(t => t.id !== this.taskId))
		},
	},
	methods: {
		async findTasks(query) {
			this.query = query
			this.foundTasks = await this.taskService.getAll({}, {s: query})
		},

		async addTaskRelation() {
			if (this.newTaskRelationTask.id === 0 && this.query !== '') {
				return this.createAndRelateTask(this.query)
			}

			if (this.newTaskRelationTask.id === 0) {
				this.$message.error({message: this.$t('task.relation.taskRequired')})
				return
			}

			const rel = new TaskRelationModel({
				taskId: this.taskId,
				otherTaskId: this.newTaskRelationTask.id,
				relationKind: this.newTaskRelationKind,
			})
			await this.taskRelationService.create(rel)
			if (!this.relatedTasks[this.newTaskRelationKind]) {
				this.relatedTasks[this.newTaskRelationKind] = []
			}
			this.relatedTasks[this.newTaskRelationKind].push(this.newTaskRelationTask)
			this.newTaskRelationTask = null
			this.saved = true
			this.showNewRelationForm = false
			setTimeout(() => {
				this.saved = false
			}, 2000)
		},

		async removeTaskRelation() {
			const rel = new TaskRelationModel({
				relationKind: this.relationToDelete.relationKind,
				taskId: this.taskId,
				otherTaskId: this.relationToDelete.otherTaskId,
			})
			try {
				await this.taskRelationService.delete(rel)

				const kind = this.relationToDelete.relationKind
				for (const t in this.relatedTasks[kind]) {
					if (this.relatedTasks[kind][t].id === this.relationToDelete.otherTaskId) {
						this.relatedTasks[kind].splice(t, 1)

						break
					}
				}

				this.saved = true
				setTimeout(() => {
					this.saved = false
				}, 2000)
			} finally {
				this.showDeleteModal = false
			}
		},

		async createAndRelateTask(title) {
			const newTask = new TaskModel({title: title, listId: this.listId})
			this.newTaskRelationTask = await this.taskService.create(newTask)
			await this.addTaskRelation()
		},

		relationKindTitle(kind, length) {
			return this.$tc(`task.relation.kinds.${kind}`, length)
		},

		mapRelatedTasks(tasks) {
			return tasks
				.map(task => {
					// by doing this here once we can save a lot of duplicate calls in the template
					const listAndNamespace = this.$store.getters['namespaces/getListAndNamespaceById'](task.listId, true)
					const {
						list,
						namespace,
					} = listAndNamespace === null ? {list: null, namespace: null} : listAndNamespace

					return {
						...task,
						differentNamespace:
							(namespace !== null &&
								namespace.id !== this.namespace.id &&
								namespace?.title) || null,
						differentList:
							(list !== null &&
								task.listId !== this.listId &&
								list?.title) || null,
					}
				})
		},
	},
})
</script>

<style lang="scss" scoped>
.add-task-relation-button {
	margin-top: -3rem;

	svg {
		transition: transform $transition;
	}

	&.is-active svg {
		transform: rotate(45deg);
	}
}

.different-list {
	color: var(--grey-500);
	width: auto;
}

.title {
	font-size: 1rem;
	margin: 0;
}

.tasks {
	padding: .5rem;
}

.task {
	display: flex;
	flex-wrap: wrap;
	justify-content: space-between;
	padding: .75rem;
	transition: background-color $transition;
	border-radius: $radius;

	&:hover {
		background-color: var(--grey-200);
	}

	a {
		color: var(--text);
		transition: color ease $transition-duration;

		&:hover {
			color: var(--grey-900);
		}
	}

	.remove {
		text-align: center;
		color: var(--danger);
		opacity: 0;
		transition: opacity $transition;
	}
}

.related-tasks:hover .tasks .task .remove {
	opacity: 1;
}

.none {
	font-style: italic;
	text-align: center;
}

:deep(.multiselect .search-results button) {
	padding: 0.5rem;
}

@include modal-transition();
</style>