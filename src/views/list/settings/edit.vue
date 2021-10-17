<template>
	<create-edit
		:title="$t('list.edit.header')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="save"
		:tertary="$t('misc.delete')"
		@tertary="$router.push({ name: 'list.list.settings.delete', params: { id: $route.params.listId } })"
	>
		<div class="field">
			<label class="label" for="title">{{ $t('list.title') }}</label>
			<div class="control">
				<input
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading || null"
					@keyup.enter="save"
					class="input"
					id="title"
					:placeholder="$t('list.edit.titlePlaceholder')"
					type="text"
					v-focus
					v-model="list.title"/>
			</div>
		</div>
		<div class="field">
			<label
				class="label"
				for="identifier"
				v-tooltip="$t('list.edit.identifierTooltip')">
				{{ $t('list.edit.identifier') }}
			</label>
			<div class="control">
				<input
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading"
					@keyup.enter="save"
					class="input"
					id="identifier"
					:placeholder="$t('list.edit.identifierPlaceholder')"
					type="text"
					v-focus
					v-model="list.identifier"/>
			</div>
		</div>
		<div class="field">
			<label class="label" for="listdescription">{{ $t('list.edit.description') }}</label>
			<div class="control">
				<editor
					:class="{ 'disabled': listService.loading}"
					:disabled="listService.loading"
					:preview-is-default="false"
					id="listdescription"
					:placeholder="$t('list.edit.descriptionPlaceholder')"
					v-model="list.description"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('list.edit.color') }}</label>
			<div class="control">
				<color-picker v-model="list.hexColor"/>
			</div>
		</div>

	</create-edit>
</template>

<script>
import AsyncEditor from '@/components/input/AsyncEditor'

import ListModel from '@/models/list'
import ListService from '@/services/list'
import ColorPicker from '@/components/input/colorPicker.vue'
import {CURRENT_LIST} from '@/store/mutation-types'
import CreateEdit from '@/components/misc/create-edit.vue'

export default {
	name: 'list-setting-edit',
	data() {
		return {
			list: ListModel,
			listService: new ListService(),
		}
	},
	components: {
		CreateEdit,
		ColorPicker,
		editor: AsyncEditor,
	},
	watch: {
		'$route.params.listId': {
			handler: 'loadList',
			immediate: true,
		},
	},
	methods: {
		async loadList() {
			const list = new ListModel({id: this.$route.params.listId})

			const loadedList = await this.listService.get(list)
			this.list = { ...loadedList }
		},

		async save() {
			await this.$store.dispatch('lists/updateList', this.list)
			await this.$store.dispatch(CURRENT_LIST, this.list)
			this.setTitle(this.$t('list.edit.title', {list: this.list.title}))
			this.$message.success({message: this.$t('list.edit.success')})
			this.$router.back()
		},
	},
}
</script>
