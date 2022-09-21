<template>
	<create-edit
		:title="$t('list.edit.header')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="save"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'list.settings.delete', params: { id: listId } })"
	>
		<div class="field">
			<label class="label" for="title">{{ $t('list.title') }}</label>
			<div class="control">
				<input
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading || undefined"
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
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading || undefined"
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
				<Editor
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					:previewIsDefault="false"
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

<script lang="ts">
export default { name: 'list-setting-edit' }
</script>

<script setup lang="ts">
import type {PropType} from 'vue'
import {useRouter} from 'vue-router'
import {useStore} from '@/store'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'
import ColorPicker from '@/components/input/colorPicker.vue'
import CreateEdit from '@/components/misc/create-edit.vue'

import {CURRENT_LIST} from '@/store/mutation-types'
import type {IList} from '@/modelTypes/IList'

import {useList} from '@/stores/lists'
import {useTitle} from '@/composables/useTitle'

const props = defineProps({
	listId: {
		type: Number as PropType<IList['id']>,
		required: true,
	},
})

const router = useRouter()
const store = useStore()

const {t} = useI18n({useScope: 'global'})

const {list, save: saveList, isLoading} = useList(props.listId)

useTitle(() => list?.title ? t('list.edit.title', {list: list.title}) : '')

async function save() {
	await saveList()
	await store.dispatch(CURRENT_LIST, {list})
	router.back()
}
</script>
