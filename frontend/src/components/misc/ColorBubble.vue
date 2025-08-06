<template>
	<span
		:style="{backgroundColor: color }"
		class="color-bubble"
	/>
</template>

<script lang="ts" setup>
import type { DataType } from 'csstype'

defineProps< {
	color: DataType.Color,
}>()
</script>

<style scoped>
.color-bubble {
	display: inline-block;
	border-radius: 100%;
	block-size: 10px;
	inline-size: 10px;
	flex-shrink: 0;
}
</style>
