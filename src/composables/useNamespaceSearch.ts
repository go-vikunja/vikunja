import {ref, computed} from 'vue'
import {useNamespaceStore} from '@/stores/namespaces'

export function useNamespaceSearch() {
	const query = ref('')
	
	const namespaceStore = useNamespaceStore()
	const namespaces = computed(() => namespaceStore.searchNamespace(query.value))
	
	function findNamespaces(newQuery: string) {
		query.value = newQuery
	}

	return {
		namespaces,
		findNamespaces,
	}
}

