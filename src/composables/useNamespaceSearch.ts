import {ref, computed} from 'vue'
import {useStore} from '@/store'

export function useNameSpaceSearch() {
	const query = ref('')
	
	const store = useStore()
	const namespaces = computed(() => store.getters['namespaces/searchNamespace'](query.value))
	
	function findNamespaces(newQuery: string) {
		query.value = newQuery
	}

	return {
		namespaces,
		findNamespaces,
	}
}

