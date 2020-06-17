<template>
	<a @click="click">
		<icon icon="sort-up" v-if="order === 'asc'"/>
		<icon icon="sort-up" v-else-if="order === 'desc'" rotation="180"/>
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
