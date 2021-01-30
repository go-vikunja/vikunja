<template>
	<create-edit title="Create a new list" @create="newList()" :create-disabled="list.title === ''">
		<div class="field">
			<label class="label" for="listTitle">List Title</label>
			<div
				:class="{ 'is-loading': listService.loading }"
				class="control"
			>
				<input
					:class="{ disabled: listService.loading }"
					@keyup.enter="newList()"
					@keyup.esc="$router.back()"
					class="input"
					placeholder="The list's title goes here..."
					type="text"
					name="listTitle"
					v-focus
					v-model="list.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && list.title === ''">
			Please specify a title.
		</p>
		<div class="field">
			<label class="label">Color</label>
			<div class="control">
				<color-picker v-model="list.hexColor" />
			</div>
		</div>
	</create-edit>
</template>

<script>
import ListService from '../../services/list'
import ListModel from '../../models/list'
import CreateEdit from '@/components/misc/create-edit'
import ColorPicker from '../../components/input/colorPicker'

export default {
	name: 'NewList',
	data() {
		return {
			showError: false,
			list: ListModel,
			listService: ListService,
		}
	},
	components: {
		CreateEdit,
		ColorPicker,
	},
	created() {
		this.list = new ListModel()
		this.listService = new ListService()
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
			this.$store
				.dispatch('lists/createList', this.list)
				.then((r) => {
					this.success(
						{ message: 'The list was successfully created.' },
						this
					)
					this.$router.push({
						name: 'list.index',
						params: { listId: r.id },
					})
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
	},
}
</script>