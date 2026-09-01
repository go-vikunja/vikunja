<script setup lang="ts">
import type {Label} from '@/client/generated'
import {useLabelStyles} from '@/composables/useLabelStyles'

defineProps<{
	label: Label
}>()

const {getLabelStyles} = useLabelStyles()
</script>

<template>
	<span
		:key="label.id"
		:style="getLabelStyles(label)"
		class="tag"
	>
		<span>{{ label.title }}</span>
	</span>
</template>
