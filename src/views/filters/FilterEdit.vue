<template>
	<create-edit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="saveSavedFilter"
		:tertary="$t('misc.delete')"
		@tertary="$router.push({ name: 'filter.settings.delete', params: { id: $route.params.listId } })"
	>
		<form @submit.prevent="saveSavedFilter()">
			<div class="field">
				<label class="label" for="title">{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading || null"
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

<script setup>
import { ref, shallowRef, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { store } from '@/store'
import { success }  from '@/message'
import { useI18n } from 'vue-i18n'

import {default as Editor} from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/create-edit.vue'
import Filters from '@/components/list/partials/filters.vue'

import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'

import {objectToSnakeCase} from '@/helpers/case'
import {getSavedFilterIdFromListId} from '@/helpers/savedFilter'

const { t } = useI18n()

function useSavedFilter(listId) {
	const filterService = shallowRef(new SavedFilterService())

	const filter = ref(new SavedFilterModel())
	const filters = computed({
		get: () => filter.value.filters,
		set(value) {
			filter.value.filters = value
		},
	})
	
	// loadSavedFilter
	watch(listId, async () => {
		// We assume the listId in the route is the pseudolist
		const savedFilterId = getSavedFilterIdFromListId(route.params.listId)

		filter.value = new SavedFilterModel({id: savedFilterId })
		const response = await filterService.value.get(filter.value)
		response.filters = objectToSnakeCase(filter.value.filters)
		filter.value = response
	}, { immediate: true })

	async function save() {
		filter.value.filters = filters.value
		const response = await filterService.value.update(filter.value)
		await store.dispatch('namespaces/loadNamespaces')
		success({message: t('filters.edit.success')})
		response.filters = objectToSnakeCase(filter.value.filters)
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
const listId =	computed(() => route.params.listId)

const {
	save,
	filter,
	filters,
	filterService,
} = useSavedFilter(listId)

const router = useRouter()
async function saveSavedFilter() {
	await save()
	router.back()
}
</script>
