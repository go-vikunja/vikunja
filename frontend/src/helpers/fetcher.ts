import {getToken} from '@/helpers/auth'

export type Method = 'GET' | 'DELETE' | 'HEAD' | 'OPTIONS' | 'POST' | 'PUT' | 'PATCH'

export interface RequestConfig {
	url: string
	method?: Method
	headers?: Record<string, string>
	params?: Record<string, unknown>
	data?: unknown
	baseURL?: string
	responseType?: 'json' | 'blob' | 'text'
	onUploadProgress?: (progress: {progress: number}) => void
	transformRequest?: (data: unknown) => unknown
}

export interface HttpResponse<T = unknown> {
	data: T
	headers: Record<string, string>
	status: number
}

export class HttpError extends Error {
	status?: number
	data?: unknown
	response?: Response
}

class InterceptorManager<T> {
	private handlers: Array<(arg: T) => Promise<T> | T> = []

	use(handler: (arg: T) => Promise<T> | T) {
		this.handlers.push(handler)
	}

	async run(arg: T): Promise<T> {
		for (const handler of this.handlers) {
			arg = await handler(arg)
		}
		return arg
	}
}

class HttpClient {
	private baseURL: string
	interceptors = {
		request: new InterceptorManager<RequestConfig>(),
		response: new InterceptorManager<HttpResponse>(),
	}

	constructor(baseURL: string) {
		this.baseURL = baseURL
	}

	private buildUrl(config: RequestConfig) {
		let url = config.url
		if (!url.startsWith('http')) {
			url = (config.baseURL ?? this.baseURL ?? '') + url
		}
		if (config.params) {
			const qs = new URLSearchParams(config.params as Record<string, string>).toString()
			if (qs) {
				url += (url.includes('?') ? '&' : '?') + qs
			}
		}
		return url
	}

	private parseHeaders(headers: Headers): Record<string, string> {
		const result: Record<string, string> = {}
		headers.forEach((v, k) => {
			result[k] = v
		})
		return result
	}

	private prepareRequestBody(data: unknown, transformRequest?: (data: unknown) => unknown): unknown {
		let body = data
		if (transformRequest) {
			body = transformRequest(body)
		}
		return body
	}

	private async parseResponseData(response: Response, responseType?: string): Promise<unknown> {
		if (responseType === 'blob') {
			return await response.blob()
		}
		
		if (responseType === 'text') {
			return await response.text()
		}
		
		try {
			return await response.json()
		} catch {
			return await response.text()
		}
	}

	private async fetchRequest(config: RequestConfig): Promise<HttpResponse> {
		const url = this.buildUrl(config)
		const init: RequestInit = {method: config.method, headers: config.headers}

		// GET and HEAD requests cannot have a body
		const method = config.method?.toUpperCase()
		if (method !== 'GET' && method !== 'HEAD') {
			const body = this.prepareRequestBody(config.data, config.transformRequest)

			if (typeof body !== 'undefined') {
				if (body instanceof FormData || body instanceof Blob) {
					init.body = body as BodyInit
				} else if (typeof body === 'string') {
					init.body = body
				} else {
					init.body = JSON.stringify(body)
					if (init.headers && !(init.headers as Record<string, string>)['Content-Type']) {
						(init.headers as Record<string, string>)['Content-Type'] = 'application/json'
					}
				}
			}
		}

		const response = await fetch(url, init)
		return {
			status: response.status,
			headers: this.parseHeaders(response.headers),
			data: await this.parseResponseData(response, config.responseType),
		}
	}

	private xhrRequest(config: RequestConfig): Promise<HttpResponse> {
		return new Promise((resolve, reject) => {
			const url = this.buildUrl(config)
			const xhr = new XMLHttpRequest()
			xhr.open(config.method ?? 'GET', url)
			if (config.headers) {
				for (const [k, v] of Object.entries(config.headers)) {
					xhr.setRequestHeader(k, v)
				}
			}
			if (config.responseType === 'blob') {
				xhr.responseType = 'blob'
			}

			xhr.upload.onprogress = ev => {
				if (config.onUploadProgress && ev.lengthComputable) {
					config.onUploadProgress({progress: ev.loaded / ev.total})
				}
			}

			xhr.onload = () => {
				const headers: Record<string, string> = {}
				const raw = xhr.getAllResponseHeaders().trim().split(/\r?\n/)
				for (const line of raw) {
					const parts = line.split(': ')
					headers[parts.shift()!.toLowerCase()] = parts.join(': ')
				}
				let data: unknown = xhr.response
				if (config.responseType === 'blob') {
					data = xhr.response
				} else {
					try {
						data = JSON.parse(xhr.responseText)
					} catch {
						data = xhr.responseText
					}
				}
				resolve({status: xhr.status, headers, data})
			}

			xhr.onerror = () => {
				const err = new HttpError('Network Error')
				reject(err)
			}

			// GET and HEAD requests cannot have a body
			const method = config.method?.toUpperCase()
			if (method === 'GET' || method === 'HEAD') {
				xhr.send()
			} else {
				const body = this.prepareRequestBody(config.data, config.transformRequest)
				if (body instanceof FormData || body instanceof Blob) {
					xhr.send(body as BodyInit)
				} else if (typeof body === 'string' || typeof body === 'undefined') {
					xhr.send(body)
				} else {
					xhr.send(JSON.stringify(body))
				}
			}
		})
	}

	async request(config: RequestConfig): Promise<HttpResponse> {
		config = await this.interceptors.request.run({...config})
		let response: HttpResponse
		if (config.onUploadProgress) {
			response = await this.xhrRequest(config)
		} else {
			response = await this.fetchRequest(config)
		}
		response = await this.interceptors.response.run(response)
		if (response.status >= 400) {
			const err = new HttpError('Request failed with status ' + response.status)
			err.status = response.status
			err.data = response.data
			throw err
		}
		return response
	}

	get(url: string, config: Partial<RequestConfig> = {}) {
		return this.request({...config, url, method: 'GET'})
	}

	delete(url: string, data?: unknown, config: Partial<RequestConfig> = {}) {
		return this.request({...config, url, method: 'DELETE', data})
	}

	post(url: string, data?: unknown, config: Partial<RequestConfig> = {}) {
		return this.request({...config, url, method: 'POST', data})
	}

	put(url: string, data?: unknown, config: Partial<RequestConfig> = {}) {
		return this.request({...config, url, method: 'PUT', data})
	}
}

export function HTTPFactory() {
	return new HttpClient(window.API_URL)
}

export function AuthenticatedHTTPFactory() {
	const instance = HTTPFactory()
	instance.interceptors.request.use(config => {
		const token = getToken()
		config.headers = {
			'Content-Type': 'application/json',
			...config.headers,
			...(token && { Authorization: `Bearer ${token}` }),
		}
		return config
	})
	return instance
}

export type FetchHttpInstance = HttpClient

export default HttpClient
