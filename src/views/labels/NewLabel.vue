<template>
	<div class="fullpage">
		<a @click="back()" class="close">
			<icon :icon="['far', 'times-circle']"/>
		</a>
		<h3>Create a new label</h3>
		<form @keyup.esc="back()" @submit.prevent="newlabel">
			<div class="field is-grouped">
				<p class="control is-expanded" v-bind:class="{ 'is-loading': labelService.loading }">
					<input
						:class="{ 'disabled': labelService.loading }"
						class="input"
						placeholder="The label title goes here..." type="text"
						v-focus
						v-model="label.title"/>
				</p>
				<p class="control">
					<button class="button is-primary has-no-shadow" type="submit">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
						Add
					</button>
				</p>
			</div>
			<p class="help is-danger" v-if="showError && label.title === ''">
				Please specify a title.
			</p>
		</form>
	</div>
</template>

<script>
import labelModel from '../../models/label'
import labelService from '../../services/label'
import {IS_FULLPAGE} from '@/store/mutation-types'
import LabelModel from '../../models/label'
import LabelService from '../../services/label'

export default {
	name: 'NewLabel',
	data() {
		return {
			labelService: labelService,
			label: labelModel,
			showError: false,
		}
	},
	created() {
		this.labelService = new LabelService()
		this.label = new LabelModel()
		this.$store.commit(IS_FULLPAGE, true)
	},
	mounted() {
		this.setTitle('Create a new label')
	},
	methods: {
		newlabel() {

			if (this.label.title === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.labelService.create(this.label)
				.then(response => {
					this.$router.push({name: 'labels.index', params: {id: response.id}})
					this.success({message: 'The label was successfully created.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		back() {
			this.$router.go(-1)
		},
	},
}
</script>
