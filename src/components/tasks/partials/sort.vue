<template>
	<a @click="click">
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
	methods: {
		click() {
			this.$emit('click')
		},
	},
}
</script>
