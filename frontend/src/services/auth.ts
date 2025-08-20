import { useConfigStore } from '@/stores/config'
import { AuthenticatedHTTPFactory, HTTPFactory } from '@/helpers/fetcher'
import { objectToSnakeCase } from '@/helpers/case'

export default class AuthService {
	private http = HTTPFactory()
	private authenticatedHttp = AuthenticatedHTTPFactory()
	private configStore = useConfigStore()

	async login(credentials) {
		const url = `${this.configStore.apiBase}/login`
		return this.http.post(url, objectToSnakeCase(credentials))
	}

	async register(credentials, language: string | null) {
		const url = `${this.configStore.apiBase}/register`
		return this.http.post(url, {
			...credentials,
			language,
		})
	}

	async refreshUserInfo() {
		const url = `${this.configStore.apiBase}/api/v1/user`
		return this.authenticatedHttp.get(url)
	}

	async verifyEmail(token: string) {
		const url = `${this.configStore.apiBase}/api/v1/user/confirm`
		return this.http.post(url, { token })
	}
}
