<template>
	<div class="fullpage">
		<a @click="back()" class="close">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new list</h3>
		<div class="field is-grouped">
			<p :class="{ 'is-loading': listService.loading}" class="control is-expanded">
				<input
					:class="{ 'disabled': listService.loading}"
					@keyup.enter="newList()"
					@keyup.esc="back()"
					class="input"
					placeholder="The list's name goes here..."
					type="text"
					v-focus
					v-model="list.title"/>
			</p>
			<p class="control">
				<button :disabled="list.title === ''" @click="newList()" class="button is-success noshadow">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
					Add
				</button>
			</p>
		</div>
		<p class="help is-danger" v-if="showError && list.title === ''">
			Please specify a title.
		</p>
	</div>
</template>

<script>
import router from '../../router'
import ListService from '../../services/list'
import ListModel from '../../models/list'
import {IS_FULLPAGE} from '@/store/mutation-types'

export default {
	name: 'NewList',
	data() {
		return {
			showError: false,
			list: ListModel,
			listService: ListService,
		}
	},
	created() {
		this.list = new ListModel()
		this.listService = new ListService()
		this.$store.commit(IS_FULLPAGE, true)
	},
	mounted() {
		this.setTitle('Create a new list')
	},
	methods: {
		newList() {
			if (this.list.title === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.list.namespaceId = this.$route.params.id
			this.$store.dispatch('lists/createList', this.list)
				.then(r => {
					this.success({message: 'The list was successfully created.'}, this)
					router.push({name: 'list.index', params: {listId: r.id}})
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		back() {
			router.go(-1)
		},
	},
}
</script>