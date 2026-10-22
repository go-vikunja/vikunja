import {createClient} from '@/client/generated/client'
import {normalizeProblemBody} from '@/client/problem'

// Cookie-only requests must skip the shared client's Authorization injection and its 401 handler,
// which calls refreshToken() — the very call some of them are.
export const publicClient = createClient({
	credentials: 'include',
	throwOnError: true,
})

publicClient.interceptors.error.use((error, response) => normalizeProblemBody(error, response))
