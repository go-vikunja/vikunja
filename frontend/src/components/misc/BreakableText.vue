<template>
	<template
		v-for="(part, i) in parts"
		:key="i"
	>
		<wbr v-if="i > 0">{{ part }}
	</template>
</template>

<script setup lang="ts">
import {computed} from 'vue'

const props = defineProps<{
	text: string
}>()

// <wbr> instead of zero-width spaces so copied emails stay intact.
const parts = computed(() => props.text.split(/(?=[@._])|(?<=\/)/))
</script>
