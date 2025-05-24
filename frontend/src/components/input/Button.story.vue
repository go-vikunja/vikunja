<script lang="ts" setup>
import {logEvent} from 'histoire/client'
import XButton from './Button.vue'
</script>

<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<Variant title="primary">
			<XButton
				variant="primary"
				@click="logEvent('Click', $event)"
			>
				Order pizza!
			</XButton>
		</Variant>

		<Variant title="secondary">
			<XButton
				variant="secondary"
				@click="logEvent('Click', $event)"
			>
				Order spaghetti!
			</XButton>
		</Variant>

		<Variant title="tertiary">
			<XButton
				variant="tertiary"
				@click="logEvent('Click', $event)"
			>
				Order tortellini!
			</XButton>
		</Variant>
	</Story>
</template>
