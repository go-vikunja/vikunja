import {createAsyncComponent} from '@/helpers/createAsyncComponent'

export default createAsyncComponent(() => import('@/components/input/editor.vue'))