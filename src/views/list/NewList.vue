<template>
	<create-edit :title="$t('list.create.header')" @create="newList()" :create-disabled="list.title === ''">
		<div class="field">
			<label class="label" for="listTitle">{{ $t('list.title') }}</label>
			<div
				:class="{ 'is-loading': listService.loading }"
				class="control"
			>
				<input
					:class="{ disabled: listService.loading }"
					@keyup.enter="newList()"
					@keyup.esc="$router.back()"
					class="input"
					:placeholder="$t('list.create.titlePlaceholder')"
					type="text"
					name="listTitle"
					v-focus
					v-model="list.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && list.title === ''">
			{{ $t('list.create.addTitleRequired') }}
		</p>
		<div class="field">
			<label class="label">{{ $t('list.color') }}</label>
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
		this.setTitle(this.$t('list.create.header'))
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
					this.success({message: this.$t('list.create.createdSuccess') })
					this.$router.push({
						name: 'list.index',
						params: { listId: r.id },
					})
				})
				.catch((e) => {
					this.error(e)
				})
		},
	},
}
</script>