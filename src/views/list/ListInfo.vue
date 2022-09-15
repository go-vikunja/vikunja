<template>
	<modal
		@close="$router.back()"
	>
		<card
			:title="list.title"
		>
			<div class="has-text-left" v-html="htmlDescription" v-if="htmlDescription !== ''"></div>
			<p v-else class="is-italic">
				{{ $t('list.noDescriptionAvailable') }}
			</p>
		</card>
	</modal>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useStore} from '@/store'
import {setupMarkdownRenderer} from '@/helpers/markdownRenderer'
import {marked} from 'marked'
import DOMPurify from 'dompurify'
import {createRandomID} from '@/helpers/randomId'

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
})

const store = useStore()
const list = computed(() => store.getters['lists/getListById'](props.listId))
const htmlDescription = computed(() => {
	const description = list.value?.description || ''
	if (description === '') {
		return ''
	}

	setupMarkdownRenderer(createRandomID())
	return DOMPurify.sanitize(marked(description), {ADD_ATTR: ['target']})
})
</script>
