<script lang="ts" setup>
import {reactive} from 'vue'
import ColorPicker from './ColorPicker.vue'

const state = reactive({
	color: '#f2f2f2',
})
</script>

<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<ColorPicker v-model="state.color" />
	</Story>
</template>
