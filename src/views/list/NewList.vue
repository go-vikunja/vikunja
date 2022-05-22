<template>
	<create-edit :title="$t('list.create.header')" @create="createNewList()" :primary-disabled="list.title === ''">
		<div class="field">
			<label class="label" for="listTitle">{{ $t('list.title') }}</label>
			<div
				:class="{ 'is-loading': listService.loading }"
				class="control"
			>
				<input
					:class="{ disabled: listService.loading }"
					@keyup.enter="createNewList()"
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

<script setup lang="ts">
import {ref, reactive, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter, useRoute} from 'vue-router'
import {useStore} from 'vuex'

import ListService from '@/services/list'
import ListModel from '@/models/list'
import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '@/components/input/colorPicker.vue'

import { success } from '@/message'
import { useTitle } from '@/composables/useTitle'

const {t} = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

useTitle(() => t('list.create.header'))

const showError = ref(false)
const list = reactive(new ListModel())
const listService = shallowReactive(new ListService())

async function createNewList() {
	if (list.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	list.namespaceId = parseInt(route.params.namespaceId)
	const newList = await store.dispatch('lists/createList', list)
	await router.push({
		name: 'list.index',
		params: { listId: newList.id },
	})
	success({message: t('list.create.createdSuccess') })
}
</script>