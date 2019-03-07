<template>
	<div class="loader-container content" :class="{ 'is-loading': labelService.loading}">
		<h1>Manage labels</h1>
		<p>
			Click on a label to edit it.
			You can edit all labels you created, you can use all lables which are associated with a task to whose list you have access.
		</p>
		<div class="columns">
			<div class="labels-list column">
				<a
					v-for="l in labels" :key="l.id"
					class="tag"
					:class="{'disabled': user.infos.id !== l.created_by.id}"
					@click="editLabel(l)"
					:style="{'background': l.hex_color, 'color': l.textColor}"
				>
					<span
						v-if="user.infos.id !== l.created_by.id"
						v-tooltip.bottom="'You are not allowed to edit this label because you dont own it.'">
						{{ l.title }}
					</span>
					<span v-else>{{ l.title }}</span>
					<a class="delete is-small" @click="deleteLabel(l)" v-if="user.infos.id === l.created_by.id"></a>

				</a>
			</div>
			<div class="column is-4" v-if="isLabelEdit">
				<div class="card">
					<header class="card-header">
						<span class="card-header-title">
							Edit Label
						</span>
						<a class="card-header-icon" @click="isTaskEdit = false">
							<span class="icon">
								<icon icon="angle-right"/>
							</span>
						</a>
					</header>
					<div class="card-content">
						<form @submit.prevent="editLabelSubmit()">
							<div class="field">
								<label class="label">Title</label>
								<div class="control">
									<input class="input" type="text" placeholder="Label title" v-model="labelEditLabel.title"/>
								</div>
							</div>
							<div class="field">
								<label class="label">Description</label>
								<div class="control">
									<textarea class="textarea" placeholder="Label description" v-model="labelEditLabel.description"></textarea>
								</div>
							</div>
							<div class="field">
								<label class="label">Color</label>
								<div class="control">
									<verte
											v-model="labelEditLabel.hex_color"
											menuPosition="top"
											picker="square"
											model="hex"
											:enableAlpha="false"
											:rgbSliders="true">
									</verte>
								</div>
							</div>
							<div class="field has-addons">
								<div class="control is-expanded">
									<button type="submit" class="button is-fullwidth is-success" :class="{ 'is-loading': labelService.loading}">
										Save
									</button>
								</div>
								<div class="control">
									<a class="button has-icon is-danger" @click="deleteLabel(labelEditLabel);isLabelEdit = false;">
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
	import verte from 'verte'
	import 'verte/dist/verte.css'

	import LabelService from '../../services/label'
	import LabelModel from '../../models/label'
	import message from '../../message'
	import auth from '../../auth'

	export default {
		name: 'ListLabels',
		components: {
			verte,
		},
		data() {
			return {
				labelService: LabelService,
				labels: [],
				labelEditLabel: LabelModel,
				isLabelEdit: false,
				user: auth.user,
			}
		},
		created() {
			this.labelService = new LabelService()
			this.labelEditLabel = new LabelModel()
			this.loadLabels()
		},
		methods: {
			loadLabels() {
				this.labelService.getAll()
					.then(r => {
						this.$set(this, 'labels', r)
					})
					.catch(e => {
						message.error(e, this)
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
						message.success({message: 'The label was successfully deleted.'}, this)
					})
					.catch(e => {
						message.error(e, this)
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
						message.success({message: 'The label was successfully updated.'}, this)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			editLabel(label) {
				if(label.created_by.id !== this.user.infos.id) {
					return
				}
				this.labelEditLabel = label
				this.isLabelEdit = true
			}
		}
	}
</script>