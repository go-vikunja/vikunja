<template>
	<div :class="{ 'is-loading': loading}" class="loader-container">
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
			<p v-else class="has-text-centered has-text-grey is-italic">
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
									:loading="loading"
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

import LabelModel from '../../models/label'
import ColorPicker from '../../components/input/colorPicker'
import LoadingComponent from '../../components/misc/loading'
import ErrorComponent from '../../components/misc/error'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

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
			labelEditLabel: LabelModel,
			isLabelEdit: false,
			editorActive: false,
		}
	},
	created() {
		this.labelEditLabel = new LabelModel()
		this.loadLabels()
	},
	mounted() {
		this.setTitle('Labels')
	},
	computed: mapState({
		userInfo: state => state.auth.info,
		labels: state => state.labels.labels,
		loading: state => state[LOADING] && state[LOADING_MODULE] === 'labels',
	}),
	methods: {
		loadLabels() {
			this.$store.dispatch('labels/loadAllLabels')
				.catch(e => {
					this.error(e)
				})
		},
		deleteLabel(label) {
			this.$store.dispatch('labels/deleteLabel', label)
				.then(() => {
					this.success({message: 'The label was successfully deleted.'})
				})
				.catch(e => {
					this.error(e)
				})
		},
		editLabelSubmit() {
			this.$store.dispatch('labels/updateLabel', this.labelEditLabel)
				.then(() => {
					this.success({message: 'The label was successfully updated.'})
				})
				.catch(e => {
					this.error(e)
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
