<template>
	<create-edit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="saveFilter"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: listId } })"
	>
		<form @submit.prevent="saveFilter()">
			<div class="field">
				<label class="label" for="title">{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading || undefined"
						@keyup.enter="saveFilter"
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
import {toRef} from 'vue'

import Editor from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/create-edit.vue'
import Filters from '@/components/list/partials/filters.vue'

import {useSavedFilter} from '@/services/savedFilter'

import type {IList} from '@/modelTypes/IList'

const props = defineProps<{ listId: IList['id'] }>()

const {
	saveFilter,
	filter,
	filters,
	filterService,
} = useSavedFilter(toRef(props, 'listId'))
</script>
