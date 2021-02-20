<template>
	<create-edit
		title="Edit This List"
		primary-icon=""
		primary-label="Save"
		@primary="save"
		tertary="Delete"
		@tertary="$router.push({ name: 'list.list.settings.delete', params: { id: $route.params.listId } })"
	>
		<div class="field">
			<label class="label" for="listtext">List Name</label>
			<div class="control">
				<input
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading"
					@keyup.enter="save"
					class="input"
					id="listtext"
					placeholder="The list title goes here..."
					type="text"
					v-focus
					v-model="list.title"/>
			</div>
		</div>
		<div class="field">
			<label
				class="label"
				for="listtext"
				v-tooltip="'The list identifier can be used to uniquely identify a task across lists. You can set it to empty to disable it.'">
				List Identifier
			</label>
			<div class="control">
				<input
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading"
					@keyup.enter="save"
					class="input"
					id="listtext"
					placeholder="The list identifier goes here..."
					type="text"
					v-focus
					v-model="list.identifier"/>
			</div>
		</div>
		<div class="field">
			<label class="label" for="listdescription">Description</label>
			<div class="control">
				<editor
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading"
					:preview-is-default="false"
					id="listdescription"
					placeholder="The lists description goes here..."
					v-model="list.description"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="list.hexColor"/>
			</div>
		</div>

	</create-edit>
</template>

<script>
import ListModel from '@/models/list'
import ListService from '@/services/list'
import ColorPicker from '@/components/input/colorPicker'
import LoadingComponent from '@/components/misc/loading'
import ErrorComponent from '@/components/misc/error'
import ListDuplicateService from '@/services/listDuplicateService'
import {CURRENT_LIST} from '@/store/mutation-types'
import CreateEdit from '@/components/misc/create-edit'

export default {
	name: 'list-setting-edit',
	data() {
		return {
			list: ListModel,
			listService: ListService,
		}
	},
	components: {
		CreateEdit,
		ColorPicker,
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '@/components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
	},
	created() {
		this.listService = new ListService()
		this.listDuplicateService = new ListDuplicateService()
		this.loadList()
	},
	methods: {
		loadList() {
			const list = new ListModel({id: this.$route.params.listId})

			this.listService.get(list)
				.then(r => {
					this.$set(this, 'list', r)
					this.$store.commit(CURRENT_LIST, r)
					this.setTitle(`Edit "${this.list.title}"`)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		save() {
			this.listService.update(this.list)
				.then(r => {
					this.$store.commit('namespaces/setListInNamespaceById', r)
					this.success({message: 'The list was successfully updated.'}, this)
					this.$router.back()
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
