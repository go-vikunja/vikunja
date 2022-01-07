import { defineAsyncComponent } from 'vue'
import ErrorComponent from '@/components/misc/error.vue'
import LoadingComponent from '@/components/misc/loading.vue'

const Editor = () => import('@/components/input/editor.vue')

export default defineAsyncComponent({
	loader: Editor,
	loadingComponent: LoadingComponent,
	errorComponent: ErrorComponent,
	timeout: 60000,
})
