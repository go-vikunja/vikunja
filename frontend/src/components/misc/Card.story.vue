<script lang="ts" setup>
import Card from './card.vue'
</script>

<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<Card>
			Card content
		</Card>
	</Story>
</template>