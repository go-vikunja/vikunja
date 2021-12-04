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

<script lang="ts" setup>
defineProps({
	to: {
		required: true,
	},
	icon: {
		type: String,
		required: false,
		default: '',
	},
})
</script>
