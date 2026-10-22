import type {Provider} from '@/client/generated'

export type IProvider = Provider & Required<Pick<Provider, 'name' | 'key' | 'auth_url' | 'client_id' | 'logout_url' | 'scope'>>
