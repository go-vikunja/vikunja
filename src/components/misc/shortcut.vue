<template>
	<span class="shortcuts">
		<template v-for="(k, i) in keys">
			<span :key="i">{{ k }}</span>
			<i v-if="i < keys.length - 1" :key="`plus${i}`">+</i>
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
