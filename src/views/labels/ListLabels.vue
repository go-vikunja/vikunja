<template>
	<div :class="{ 'is-loading': labelService.loading}" class="loader-container">
		<x-button
			:to="{name:'labels.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			New label
		</x-button>

		<div class="content">
			<h1>Manage labels</h1>
			<p v-if="labels.length > 0">
				Click on a label to edit it.
				You can edit all labels you created, you can use all labels which are associated with a task to whose
				list you have access.
			</p>
			<p v-else class="has-text-centered has-text-grey">
				You currently do not have any labels.
				<router-link :to="{name:'labels.create'}">Create a new label.</router-link>
			</p>
		</div>

		<div class="columns">
			<div class="labels-list column">
				<span
					:class="{'disabled': userInfo.id !== l.createdBy.id}" :key="l.id"
					:style="{'background': l.hexColor, 'color': l.textColor}"
					class="tag"
					v-for="l in labels"
				>
					<span
						v-if="userInfo.id !== l.createdBy.id"
						v-tooltip.bottom="'You are not allowed to edit this label because you dont own it.'">
						{{ l.title }}
					</span>
					<a
						:style="{'color': l.textColor}"
						@click="editLabel(l)"
						v-else>
						{{ l.title }}
					</a>
					<a @click="deleteLabel(l)" class="delete is-small" v-if="userInfo.id === l.createdBy.id"></a>
				</span>
			</div>
			<div class="column is-4" v-if="isLabelEdit">
				<card title="Edit Label" :has-close="true" @close="() => isLabelEdit = false">
					<form @submit.prevent="editLabelSubmit()">
						<div class="field">
							<label class="label">Title</label>
							<div class="control">
								<input
									class="input"
									placeholder="Label title"
									type="text"
									v-model="labelEditLabel.title"/>
							</div>
						</div>
						<div class="field">
							<label class="label">Description</label>
							<div class="control">
								<editor
									:preview-is-default="false"
									placeholder="Label description"
									v-if="editorActive"
									v-model="labelEditLabel.description"
								/>
							</div>
						</div>
						<div class="field">
							<label class="label">Color</label>
							<div class="control">
								<color-picker v-model="labelEditLabel.hexColor"/>
							</div>
						</div>
						<div class="field has-addons">
							<div class="control is-expanded">
								<x-button
									:loading="labelService.loading"
									class="is-fullwidth"
									@click="editLabelSubmit()"
								>
									Save
								</x-button>
							</div>
							<div class="control">
								<x-button
									@click="() => {deleteLabel(labelEditLabel);isLabelEdit = false}"
									icon="trash-alt"
									class="is-danger"
								/>
							</div>
						</div>
					</form>
				</card>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'

import LabelService from '../../services/label'
import LabelModel from '../../models/label'
import ColorPicker from '../../components/input/colorPicker'
import LoadingComponent from '../../components/misc/loading'
import ErrorComponent from '../../components/misc/error'

export default {
	name: 'ListLabels',
	components: {
		ColorPicker,
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '../../components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	data() {
		return {
			labelService: LabelService,
			labels: [],
			labelEditLabel: LabelModel,
			isLabelEdit: false,
			editorActive: false,
		}
	},
	created() {
		this.labelService = new LabelService()
		this.labelEditLabel = new LabelModel()
		this.loadLabels()
	},
	mounted() {
		this.setTitle('Labels')
	},
	computed: mapState({
		userInfo: state => state.auth.info,
	}),
	methods: {
		loadLabels() {
			const getAllLabels = (page = 1) => {
				return this.labelService.getAll({}, {}, page)
					.then(labels => {
						if (page < this.labelService.totalPages) {
							return getAllLabels(page + 1)
								.then(nextLabels => {
									return labels.concat(nextLabels)
								})
						} else {
							return labels
						}
					})
					.catch(e => {
						return Promise.reject(e)
					})
			}

			getAllLabels()
				.then(r => {
					this.$set(this, 'labels', r)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		deleteLabel(label) {
			this.labelService.delete(label)
				.then(() => {
					// Remove the label from the list
					for (const l in this.labels) {
						if (this.labels[l].id === label.id) {
							this.labels.splice(l, 1)
						}
					}
					this.success({message: 'The label was successfully deleted.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		editLabelSubmit() {
			this.labelService.update(this.labelEditLabel)
				.then(r => {
					for (const l in this.labels) {
						if (this.labels[l].id === r.id) {
							this.$set(this.labels, l, r)
						}
					}
					this.success({message: 'The label was successfully updated.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		editLabel(label) {
			if (label.createdBy.id !== this.userInfo.id) {
				return
			}
			this.labelEditLabel = label
			this.isLabelEdit = true

			// This makes the editor trigger its mounted function again which makes it forget every input
			// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
			// which made it impossible to detect change from the outside. Therefore the component would
			// not update if new content from the outside was made available.
			// See https://github.com/NikulinIlya/vue-easymde/issues/3
			this.editorActive = false
			this.$nextTick(() => this.editorActive = true)
		},
	},
}
</script>
