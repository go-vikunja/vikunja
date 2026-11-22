<script setup lang="ts">
import {ref} from 'vue'
import EmojiPicker from './EmojiPicker.vue'

const selected = ref('')
</script>

<template>
	<Story title="Emoji picker">
		<EmojiPicker @select="selected = $event" />
		<p v-if="selected">
			{{ selected }}
		</p>
	</Story>
</template>
