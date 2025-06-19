import { defineAsyncComponent, type AsyncComponentLoader, type AsyncComponentOptions, type Component, type ComponentPublicInstance } from 'vue'

import ErrorComponent from '@/components/misc/Error.vue'
import LoadingComponent from '@/components/misc/Loading.vue'

const DEFAULT_TIMEOUT = 60000

export function createAsyncComponent<T extends Component = {
	new (): ComponentPublicInstance;
}>(source: AsyncComponentLoader<T> | AsyncComponentOptions<T>): T {
	if (typeof source === 'function') {
		source = { loader: source }
	}

	return defineAsyncComponent({
		...source,
		loadingComponent: LoadingComponent,
		errorComponent: ErrorComponent,
		timeout: DEFAULT_TIMEOUT,
	})
}
