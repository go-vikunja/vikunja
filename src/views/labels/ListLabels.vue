<template>
	<div :class="{ 'is-loading': loading}" class="loader-container">
		<x-button
			:to="{name:'labels.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			{{ $t('label.create.header') }}
		</x-button>

		<div class="content">
			<h1>{{ $t('label.manage') }}</h1>
			<p v-if="Object.entries(labels).length > 0">
				{{ $t('label.description') }}
			</p>
			<p v-else class="has-text-centered has-text-grey is-italic">
				{{ $t('label.newCTA') }}
				<router-link :to="{name:'labels.create'}">{{ $t('label.create.title') }}.</router-link>
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
						v-tooltip.bottom="$t('label.edit.forbidden')">
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
				<card :title="$t('label.edit.header')" :has-close="true" @close="() => isLabelEdit = false">
					<form @submit.prevent="editLabelSubmit()">
						<div class="field">
							<label class="label">{{ $t('label.attributes.title') }}</label>
							<div class="control">
								<input
									class="input"
									:placeholder="$t('label.attributes.titlePlaceholder')"
									type="text"
									v-model="labelEditLabel.title"/>
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('label.attributes.description') }}</label>
							<div class="control">
								<editor
									:preview-is-default="false"
									:placeholder="$t('label.attributes.description')"
									v-if="editorActive"
									v-model="labelEditLabel.description"
								/>
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('label.attributes.color') }}</label>
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
									{{ $t('misc.save') }}
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
		this.setTitle(this.$t('label.title'))
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
					this.success({message: this.$t('label.deleteSuccess')})
				})
				.catch(e => {
					this.error(e)
				})
		},
		editLabelSubmit() {
			this.$store.dispatch('labels/updateLabel', this.labelEditLabel)
				.then(() => {
					this.success({message: this.$t('label.edit.success')})
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
