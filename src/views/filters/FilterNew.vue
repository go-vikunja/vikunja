<template>
	<modal
		@close="$router.back()"
		variant="hint-modal"
	>
				<card class="has-no-shadow" :title="$t('filters.create.title')">
					<p>
						{{ $t('filters.create.description') }}
					</p>
					<div class="field">
						<label class="label" for="title">{{ $t('filters.attributes.title') }}</label>
						<div class="control">
							<input
								v-model="savedFilter.title"
								:class="{ 'disabled': savedFilterService.loading}"
								:disabled="savedFilterService.loading || undefined"
								class="input"
								id="Title"
								:placeholder="$t('filters.attributes.titlePlaceholder')"
								type="text"
								v-focus
							/>
						</div>
					</div>
					<div class="field">
						<label class="label" for="description">{{ $t('filters.attributes.description') }}</label>
						<div class="control">
							<editor
								:key="savedFilter.id"
								v-model="savedFilter.description"
								:class="{ 'disabled': savedFilterService.loading}"
								:disabled="savedFilterService.loading"
								:preview-is-default="false"
								id="description"
								:placeholder="$t('filters.attributes.descriptionPlaceholder')"
							/>
						</div>
					</div>
					<div class="field">
						<label class="label" for="filters">{{ $t('filters.title') }}</label>
						<div class="control">
							<Filters
								:class="{ 'disabled': savedFilterService.loading}"
								:disabled="savedFilterService.loading"
								class="has-no-shadow has-no-border"
								v-model="filters"
							/>
						</div>
					</div>
					<x-button
						:loading="savedFilterService.loading"
						:disabled="savedFilterService.loading"
						@click="create()"
						class="is-fullwidth"
					>
						{{ $t('filters.create.action') }}
					</x-button>
				</card>
	</modal>
</template>

<script setup lang="ts">
import { ref, shallowRef, computed } from 'vue'

import { store } from '@/store'
import { useRouter } from 'vue-router'

import {default as Editor} from '@/components/input/AsyncEditor'
import Filters from '@/components/list/partials/filters.vue'

import SavedFilterService from '@/services/savedFilter'
import SavedFilterModel from '@/models/savedFilter'

const savedFilterService = shallowRef(new SavedFilterService())

const savedFilter = ref(new SavedFilterModel())
const filters = computed({
	get: () => savedFilter.value.filters,
	set: (value) => (savedFilter.value.filters = value),
})

const router = useRouter()
async function create() {
	savedFilter.value = await savedFilterService.value.create(savedFilter.value)
	await store.dispatch('namespaces/loadNamespaces')
	router.push({name: 'list.index', params: {listId: savedFilter.value.getListId()}})
}
</script>
