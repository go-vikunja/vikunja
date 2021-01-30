<template>
	<router-link
		:to="to"
		class="dropdown-item">
		<span class="icon" v-if="icon !== ''">
			<icon :icon="icon"/>
		</span>
		<span>
			<slot></slot>
		</span>
	</router-link>
</template>

<script>
export default {
	name: 'dropdown-item',
	props: {
		to: {
			required: true,
		},
		icon: {
			type: String,
			required: false,
			default: '',
		}
	},
}
</script>
