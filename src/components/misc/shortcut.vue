<template>
	<component :is="is" class="shortcuts">
		<template v-for="(k, i) in keys" :key="i">
			<kbd>{{ k }}</kbd>
			<span v-if="i < keys.length - 1">{{ combination }}</span>
		</template>
	</component>
</template>

<script>
export default {
	name: 'shortcut',
	props: {
		keys: {
			type: Array,
			required: true,
		},
		combination: {
			type: String,
			default: '+',
		},
		is: {
			type: String,
			default: 'div',
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
	border: 1px solid var(--grey-300);
	background: var(--grey-100);
	border-radius: 3px;
	font-size: .75rem;
}

span {
	padding: 0 .25rem;
}
</style>