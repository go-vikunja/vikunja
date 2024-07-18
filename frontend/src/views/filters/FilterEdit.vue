<template>
	<CreateEdit
		:title="$t('filters.edit.title')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:tertiary="$t('misc.delete')"
		@primary="saveFilterWithValidation"
		@tertiary="$router.push({ name: 'filter.settings.delete', params: { id: projectId } })"
	>
		<form @submit.prevent="saveFilterWithValidation()">
			<div class="field">
				<label
					class="label"
					for="title"
				>{{ $t('filters.attributes.title') }}</label>
				<div class="control">
					<input
						id="Title"
						v-model="filter.title"
						v-focus
						:class="{ 'disabled': filterService.loading, 'is-danger': !titleValid }"
						:disabled="filterService.loading || undefined"
						class="input"
						:placeholder="$t('filters.attributes.titlePlaceholder')"
						type="text"
						@focusout="validateTitleField"
					>
				</div>
				<p
					v-if="!titleValid"
					class="help is-danger"
				>
					{{ $t('filters.create.titleRequired') }}
				</p>
			</div>
			<div class="field">
				<label
					class="label"
					for="description"
				>{{ $t('filters.attributes.description') }}</label>
				<div class="control">
					<Editor
						id="description"
						v-model="filter.description"
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						:placeholder="$t('filters.attributes.descriptionPlaceholder')"
					/>
				</div>
			</div>
			<div class="field">
				<label
					class="label"
					for="filters"
				>{{ $t('filters.title') }}</label>
				<div class="control">
					<Filters
						v-model="filters"
						:class="{ 'disabled': filterService.loading}"
						:disabled="filterService.loading"
						class="has-no-shadow has-no-border"
						:has-footer="false"
						:change-immediately="true"
					/>
				</div>
			</div>
		</form>
	</CreateEdit>
</template>

<script setup lang="ts">
import Editor from '@/components/input/AsyncEditor'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import Filters from '@/components/project/partials/Filters.vue'

import {useSavedFilter} from '@/services/savedFilter'

import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	projectId: IProject['id'],
}>()

const {
	saveFilterWithValidation,
	filter,
	filters,
	filterService,
	titleValid,
	validateTitleField,
} = useSavedFilter(() => props.projectId)
</script>
