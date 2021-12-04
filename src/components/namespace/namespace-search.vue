<template>
	<multiselect
		:placeholder="$t('namespace.search')"
		@search="findNamespaces"
		:search-results="namespaces"
		@select="select"
		label="title"
		:search-delay="10"
	/>
</template>

<script lang="ts" setup>
import {ref, computed} from 'vue'
import {useStore} from 'vuex'
import Multiselect from '@/components/input/multiselect.vue'

const emit = defineEmits(['selected'])

const query = ref('')

const store = useStore()
const namespaces = computed(() => store.getters['namespaces/searchNamespace'](query.value))

function findNamespaces(newQuery: string) {
	query.value = newQuery
}

function select(namespace) {
	emit('selected', namespace)
}
</script>
