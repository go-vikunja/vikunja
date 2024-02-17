<script setup lang="ts">
import FilterInput from '@/components/project/partials/FilterInput.vue'
import {ref} from "vue";

function initState(value: string) {
	return {
		value,
	}
}

const value = ref('')
</script>

<template>
	<Story title="Filter Input">
		<FilterInput v-model="value"/>
	</Story>
</template>
