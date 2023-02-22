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
								v-model="filter.title"
								:class="{ 'disabled': filterService.loading, 'is-danger': !titleValid  }"
								:disabled="filterService.loading || undefined"
								class="input"
								id="Title"
								:placeholder="$t('filters.attributes.titlePlaceholder')"
								type="text"
								v-focus
								@focusout="validateTitleField"
							/>
						</div>
						<p class="help is-danger" v-if="!titleValid">{{ $t('filters.create.titleRequired') }}</p>
					</div>
					<div class="field">
						<label class="label" for="description">{{ $t('filters.attributes.description') }}</label>
						<div class="control">
							<editor
								:key="filter.id"
								v-model="filter.description"
								:class="{ 'disabled': filterService.loading}"
								:disabled="filterService.loading"
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
								:class="{ 'disabled': filterService.loading}"
								:disabled="filterService.loading"
								class="has-no-shadow has-no-border"
								v-model="filters"
							/>
						</div>
					</div>

					<template #footer>
						<x-button
							:loading="filterService.loading"
							:disabled="filterService.loading || !titleValid"
							@click="createFilterWithValidation()"
							class="is-fullwidth"
						>
							{{ $t('filters.create.action') }}
						</x-button>
					</template>
				</card>
	</modal>
</template>

<script setup lang="ts">
import Editor from '@/components/input/AsyncEditor'
import Filters from '@/components/list/partials/filters.vue'

import {useSavedFilter} from '@/services/savedFilter'

const {
	filter,
	filters,
	createFilterWithValidation,
	filterService,
	titleValid,
	validateTitleField,
} = useSavedFilter()
</script>
