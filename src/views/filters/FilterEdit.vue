<template>
	<create-edit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="saveSavedFilter"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: $route.params.listId } })"
	>
		<form @submit.prevent="saveSavedFilter()">
			<div class="field">
				<label class="label" for="title">{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading || undefined"
						@keyup.enter="saveSavedFilter"
						class="input"
						id="title"
						:placeholder="$t('filters.attributes.titlePlaceholder')"
						type="text"
						v-focus
						v-model="filter.title"/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="description">{{ $t('filters.attributes.description') }}</label>
				<div class="control">
					<editor
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						:preview-is-default="false"
						id="description"
						:placeholder="$t('filters.attributes.descriptionPlaceholder')"
						v-model="filter.description"
					/>
				</div>
			</div>
			<div class="field">
				<label class="label" for="filters">{{ $t('filters.title') }}</label>
				<div class="control">
					<Filters
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						class="has-no-shadow has-no-border"
						v-model="filters"
					/>
				</div>
			</div>
		</form>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, shallowRef, computed, watch, unref } from 'vue'
import {useRouter, useRoute} from 'vue-router'
import {useStore} from '@/store'
import {success} from '@/message'
import {useI18n} from 'vue-i18n'
import type {MaybeRef} from '@vueuse/core'
import {CURRENT_LIST} from '@/store/mutation-types'

import {default as Editor} from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/create-edit.vue'
import Filters from '@/components/list/partials/filters.vue'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'

import {objectToSnakeCase} from '@/helpers/case'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'
import type {IList} from '@/modelTypes/IList'
import {useNamespaceStore} from '@/stores/namespaces'

const {t} = useI18n({useScope: 'global'})
const namespaceStore = useNamespaceStore()

function useSavedFilter(listId: MaybeRef<IList['id']>) {
	const filterService = shallowRef(new SavedFilterService())

	const filter = ref(new SavedFilterModel())
	const filters = computed({
		get: () => filter.value.filters,
		set(value) {
			filter.value.filters = value
		},
	})

	// loadSavedFilter
	watch(() => unref(listId), async () => {
		// We assume the listId in the route is the pseudolist
		const savedFilterId = getSavedFilterIdFromListId(Number(route.params.listId as string))

		filter.value = new SavedFilterModel({id: savedFilterId})
		const response = await filterService.value.get(filter.value)
		response.filters = objectToSnakeCase(response.filters)
		filter.value = response
	}, {immediate: true})

	async function save() {
		filter.value.filters = filters.value
		const response = await filterService.value.update(filter.value)
		await namespaceStore.loadNamespaces()
		success({message: t('filters.edit.success')})
		response.filters = objectToSnakeCase(response.filters)
		filter.value = response
	}

	return {
		save,
		filter,
		filters,
		filterService,
	}
}

const route = useRoute()
const store = useStore()
const listId =	computed(() => Number(route.params.listId as string))

const {
	save,
	filter,
	filters,
	filterService,
} = useSavedFilter(listId)

const router = useRouter()

async function saveSavedFilter() {
	await save()
	await store.dispatch(CURRENT_LIST, {list: filter})
	router.back()
}
</script>
