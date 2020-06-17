<template>
	<div class="loader-container content" :class="{ 'is-loading': labelService.loading}">
		<h1>Manage labels</h1>
		<p>
			Click on a label to edit it.
			You can edit all labels you created, you can use all labels which are associated with a task to whose list
			you have access.
		</p>
		<div class="columns">
			<div class="labels-list column">
				<span
						v-for="l in labels" :key="l.id"
						class="tag"
						:class="{'disabled': userInfo.id !== l.createdBy.id}"
						:style="{'background': l.hexColor, 'color': l.textColor}"
				>
					<span
							v-if="userInfo.id !== l.createdBy.id"
							v-tooltip.bottom="'You are not allowed to edit this label because you dont own it.'">
						{{ l.title }}
					</span>
					<a
							@click="editLabel(l)"
							:style="{'color': l.textColor}"
							v-else>
						{{ l.title }}
					</a>
					<a class="delete is-small" @click="deleteLabel(l)" v-if="userInfo.id === l.createdBy.id"></a>
				</span>
			</div>
			<div class="column is-4" v-if="isLabelEdit">
				<div class="card">
					<header class="card-header">
						<span class="card-header-title">
							Edit Label
						</span>
						<a class="card-header-icon" @click="isLabelEdit = false">
							<span class="icon">
								<icon icon="times"/>
							</span>
						</a>
					</header>
					<div class="card-content">
						<form @submit.prevent="editLabelSubmit()">
							<div class="field">
								<label class="label">Title</label>
								<div class="control">
									<input
											class="input"
											type="text"
											placeholder="Label title"
											v-model="labelEditLabel.title"/>
								</div>
							</div>
							<div class="field">
								<label class="label">Description</label>
								<div class="control">
									<textarea
											class="textarea"
											placeholder="Label description"
											v-model="labelEditLabel.description"></textarea>
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
									<button type="submit" class="button is-fullwidth is-success"
											:class="{ 'is-loading': labelService.loading}">
										Save
									</button>
								</div>
								<div class="control">
									<a
											class="button has-icon is-danger"
											@click="() => {deleteLabel(labelEditLabel);isLabelEdit = false}">
										<span class="icon">
											<icon icon="trash-alt"/>
										</span>
									</a>
								</div>
							</div>
						</form>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import {mapState} from 'vuex'

	import LabelService from '../../services/label'
	import LabelModel from '../../models/label'
	import ColorPicker from '../../components/input/colorPicker'

	export default {
		name: 'ListLabels',
		components: {
			ColorPicker,
		},
		data() {
			return {
				labelService: LabelService,
				labels: [],
				labelEditLabel: LabelModel,
				isLabelEdit: false,
			}
		},
		created() {
			this.labelService = new LabelService()
			this.labelEditLabel = new LabelModel()
			this.loadLabels()
		},
		computed: mapState({
			userInfo: state => state.auth.info
		}),
		methods: {
			loadLabels() {
				const getAllLabels = (page = 1) => {
					return this.labelService.getAll({}, {}, page)
						.then(labels => {
							if(page < this.labelService.totalPages) {
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
			}
		}
	}
</script>
