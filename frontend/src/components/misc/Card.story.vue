<script lang="ts" setup>
import Card from './Card.vue'
</script>

<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<Card>
			Card content
		</Card>
	</Story>
</template>
