<template>
	<span class="shortcuts">
		<template v-for="(k, i) in keys" :key="i">
			<kbd>{{ k }}</kbd>
			<span v-if="i < keys.length - 1">+</span>
		</template>
	</span>
</template>

<script>
export default {
	name: 'shortcut',
	props: {
		keys: {
			type: Array,
			required: true,
		},
	},
}
</script>

<style lang="scss" scoped>
.shortcuts {
	display: flex;
	align-items: center;
}

kbd {
	padding: .1rem .35rem;
	border: 1px solid $grey-300;
	background: $grey-100;
	border-radius: 3px;
	font-size: .75rem;
}

span {
	padding: 0 .25rem;
}
</style>