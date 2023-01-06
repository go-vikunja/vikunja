<script lang="ts" setup>
import {logEvent} from 'histoire/client'
import XButton from './button.vue'
</script>

<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<Variant title="primary">
			<XButton @click="logEvent('Click', $event)" variant="primary">
				Order pizza!
			</XButton>
		</Variant>

		<Variant title="secondary">
			<XButton @click="logEvent('Click', $event)" variant="secondary">
				Order spaghetti!
			</XButton>
		</Variant>

		<Variant title="tertiary">
			<XButton @click="logEvent('Click', $event)" variant="tertiary">
				Order tortellini!
			</XButton>
		</Variant>
	</Story>
</template>