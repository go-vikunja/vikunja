<template>
	<a @click="$emit('click')">
		<icon icon="sort-up" v-if="order === 'asc'"/>
		<icon icon="sort-up" rotation="180" v-else-if="order === 'desc'"/>
		<icon icon="sort" v-else/>
	</a>
</template>

<script>
export default {
	name: 'sort',
	props: {
		order: {
			type: String,
			default: 'none',
		},
	},
	emits: ['click'],
}
</script>
