<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import Multiselect from '@/components/input/Multiselect.vue'
import {useLabels} from '@/composables/useLabels'

const ids = defineModel<number[]>({required: true})
const {t} = useI18n({useScope: 'global'})
const {labels, isPending} = useLabels()
const query = ref('')
const selected = computed(() => ids.value.map(id => ({
	id,
	title: labels.value.find(label => label.id === id)?.title ?? t('filters.simple.unknownLabel', {id}),
})))
const results = computed(() => labels.value.flatMap(label => label.id !== undefined && !ids.value.includes(label.id) &&
	(label.title ?? '').toLowerCase().includes(query.value.toLowerCase())
	? [{
		id: label.id,
		title: label.title ?? '',
	}] : []))
</script>

<template>
	<Multiselect
		:model-value="selected"
		:search-results="results"
		:loading="isPending"
		:multiple="true"
		:show-empty="true"
		:close-after-select="false"
		:aria-label="t('task.attributes.labels')"
		:placeholder="t('task.label.placeholder')"
		label="title"
		@search="query = $event"
		@update:modelValue="ids = Array.isArray($event) ? $event.map(label => label.id) : []"
	/>
</template>
