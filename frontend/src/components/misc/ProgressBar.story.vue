<script setup lang="ts">
import {ref} from 'vue'

import ProgressBar from './ProgressBar.vue'

const value = ref(50)
</script>

<template>
	<Story>
		<Variant title="Default">
			<ProgressBar :value="value" />
		</Variant>
	</Story>
</template>
