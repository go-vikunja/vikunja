<template>
	<create-edit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="saveFilterWithValidation"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: listId } })"
	>
		<form @submit.prevent="saveFilterWithValidation()">
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
	saveFilterWithValidation,
	filter,
	filters,
	filterService,
	titleValid,
	validateTitleField,
} = useSavedFilter(toRef(props, 'listId'))
</script>
