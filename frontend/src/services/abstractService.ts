import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import type {Method} from 'axios'

import {objectToSnakeCase} from '@/helpers/case'
import AbstractModel from '@/models/abstractModel'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {Permission} from '@/constants/permissions'

interface Paths {
	create : string
	get : string
	getAll : string
	update : string
	delete : string
	reset?: string
}

function convertObject(o: Record<string, unknown>) {
	if (o instanceof Date) {
		return o.toISOString()
	}

	return o
}

function prepareParams(params: Record<string, unknown | unknown[]>) {
	if (typeof params !== 'object') {
		return params
	}

	for (const p in params) {
		if (Array.isArray(params[p])) {
			params[p] = params[p].map(convertObject)
			continue
		}

		params[p] = convertObject(params[p])
	}

	return objectToSnakeCase(params)
}

export default abstract class AbstractService<Model extends IAbstract = IAbstract> {

	/////////////////////////////
	// Initial variable definitions
	///////////////////////////

	http
	loading = false
	uploadProgress = 0
	paths: Paths = {
		create: '',
		get: '',
		getAll: '',
		update: '',
		delete: '',
	}
	// This contains the total number of pages and the number of results for the current page
	totalPages = 0
	resultCount = 0

	/////////////
	// Service init
	///////////

	/**
	 * The abstract constructor.
	 * @param [paths] An object with all paths.
	 */
	constructor(paths : Partial<Paths> = {}) {
		this.http = AuthenticatedHTTPFactory()

		// Set the interceptors to process every request
		this.http.interceptors.request.use((config) => {
			switch (config.method) {
				case 'post':
					if (this.useUpdateInterceptor()) {
						config.data = this.beforeUpdate(config.data)
						if(this.autoTransformBeforePost()) {
							config.data = objectToSnakeCase(config.data)
						}
					}
					break
				case 'put':
					if (this.useCreateInterceptor()) {
						config.data = this.beforeCreate(config.data)
						if(this.autoTransformBeforePut()) {
							config.data = objectToSnakeCase(config.data)
						}
					}
					break
				case 'delete':
					if (this.useDeleteInterceptor()) {
						config.data = this.beforeDelete(config.data)
						if(this.autoTransformBeforeDelete()) {
							config.data = objectToSnakeCase(config.data)
						}
					}
					break
			}
			return config
		})

		Object.assign(this.paths, paths)
	}

	/**
	 * Whether or not to use the create interceptor which processes a request payload into json
	 */
	useCreateInterceptor(): boolean {
		return true
	}

	/**
	 * Whether or not to use the update interceptor which processes a request payload into json
	 */
	useUpdateInterceptor(): boolean {
		return true
	}

	/**
	 * Whether or not to use the delete interceptor which processes a request payload into json
	 */
	useDeleteInterceptor(): boolean {
		return true
	}
	
	autoTransformBeforeSend(): boolean {
		return true
	}

	autoTransformBeforePost(): boolean {
		return this.autoTransformBeforeSend()
	}

	autoTransformBeforePut(): boolean {
		return this.autoTransformBeforeSend()
	}

	autoTransformBeforeDelete(): boolean {
		return this.autoTransformBeforeSend()
	}

	/////////////////
	// Helper functions
	///////////////

	/**
	 * Returns an object with all route parameters and their values.
	 * @example
	 * getRouteReplacements(
	 * 	'/tasks/{taskId}/assignees/{userId}',
	 * 	{ taskId: 7, userId: 2 },
	 * )
	 * // { "{taskId}": 7, "{userId}": 2 }
	 */
	getRouteReplacements(route : string, parameters : Record<string, unknown> = {}) {
		const replace$$1: Record<string, unknown> = {}
		let pattern = this.getRouteParameterPattern()
		pattern = new RegExp(pattern instanceof RegExp ? pattern.source : pattern, 'g')

		for (let parameter; (parameter = pattern.exec(route)) !== null;) {
			replace$$1[parameter[0]] = parameters[parameter[1]]
		}

		return replace$$1
	}

	/**
	 * Holds the replacement pattern for url paths, can be overwritten by implementations.
	 */
	getRouteParameterPattern(): RegExp {
		return /{([^}]+)}/
	}

	/**
	 * Returns a fully-ready-ready-to-make-a-request-to route with replaced parameters.
	 * @example
	 * getReplacedRoute('/projects/{projectId}/tasks', { projectId: 3 })
	 * === '/projects/{projectId}/tasks'
	 */
	getReplacedRoute(path : string, pathparams : Record<string, unknown>) : string {
		const replacements = this.getRouteReplacements(path, pathparams)
		return Object.entries(replacements).reduce(
			(result, [parameter, value]) => result.replace(parameter, value as string),
			path,
		)
	}

	/**
	 * setLoading is a method which sets the loading variable to true, after a timeout of 100ms.
	 * It has the timeout to prevent the loading indicator from showing for only a blink of an eye in the
	 * case the api returns a response in < 100ms.
	 * But because the timeout is created using setTimeout, it will still trigger even if the request is
	 * already finished, so we return a method to call in that case.
	 */
	setLoading() {
		const timeout = setTimeout(() => {
			this.loading = true
		}, 100)
		return () => {
			clearTimeout(timeout)
			this.loading = false
		}
	}

	//////////////////
	// Default factories
	// It is possible to specify a factory for each type of request.
	// This makes it possible to have different models returned from different routes.
	// Specific factories for each request are completely optional, if these are not specified, the default factory is used.
	////////////////

	/**
	 * The modelFactory returns a model from an object.
	 * This one here is the default one, usually the service definitions for a model will override this.
	 */
	modelFactory(data : Partial<Model>) {
		return data as Model
	}

	/**
	 * This is the model factory for get requests.
	 */
	modelGetFactory(data : Partial<Model>) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for get all requests.
	 */
	modelGetAllFactory(data : Partial<Model>) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for create requests.
	 */
	modelCreateFactory(data : Partial<Model>) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for update requests.
	 */
	modelUpdateFactory(data : Partial<Model>) {
		return this.modelFactory(data)
	}

	//////////////
	// Preprocessors
	////////////

	/**
	 * Default preprocessor for get requests
	 */
	beforeGet(model : Model) {
		return model
	}

	/**
	 * Default preprocessor for create requests
	 */
	beforeCreate(model : Model) {
		return model
	}

	/**
	 * Default preprocessor for update requests
	 */
	beforeUpdate(model : Model) {
		return model
	}

	/**
	 * Default preprocessor for delete requests
	 */
	beforeDelete(model : Model) {
		return model
	}

	///////////////
	// Global actions
	/////////////

	/**
	 * Performs a get request to the url specified before.
	 * @param model The model to use. The request path is built using the values from the model.
	 * @param params Optional query parameters
	 */
	get(model : Model, params = {}) {
		if (this.paths.get === '') {
			throw new Error('This model is not able to get data.')
		}

		return this.getM(this.paths.get, model, params)
	}

	/**
	 * This is a more abstract implementation which only does a get request.
	 * Services which need more flexibility can use this.
	 */
	async getM(url : string, model : Model = new AbstractModel({}), params: Record<string, unknown> = {}) {
		const cancel = this.setLoading()

		model = this.beforeGet(model)
		const finalUrl = this.getReplacedRoute(url, model)

		try {
			const response = await this.http.get(finalUrl, {params: prepareParams(params)})
			const result = this.modelGetFactory(response.data)
			result.maxPermission = Number(response.headers['x-max-permission']) as Permission
			return result
		} finally {
			cancel()
		}
	}

	async getBlobUrl(url : string, method : Method = 'GET', data = {}) {
		const response = await this.http({
			url,
			method,
			responseType: 'blob',
			data,
		})
		
		// Handle SVG blobs specially - convert to data URL for better browser compatibility
		if (response.data.type === 'image/svg+xml') {
			return new Promise((resolve, reject) => {
				const reader = new FileReader()
				reader.onload = () => resolve(reader.result as string)
				reader.onerror = reject
				reader.readAsDataURL(response.data)
			})
		}
		
		return window.URL.createObjectURL(new Blob([response.data]))
	}

	/**
	 * Performs a get request to the url specified before.
	 * The difference between this and get() is this one is used to get a bunch of data (an array), not just a single object.
	 * @param model The model to use. The request path is built using the values from the model.
	 * @param params Optional query parameters
	 * @param page The page to get
	 */
	async getAll(model : Model = new AbstractModel({}), params = {}, page = 1): Promise<Model[]> {
		if (this.paths.getAll === '') {
			throw new Error('This model is not able to get data.')
		}

		params.page = page

		const cancel = this.setLoading()
		model = this.beforeGet(model)
		const finalUrl = this.getReplacedRoute(this.paths.getAll, model)

		try {
			const response = await this.http.get(finalUrl, {params: prepareParams(params)})
			this.resultCount = Number(response.headers['x-pagination-result-count'])
			this.totalPages = Number(response.headers['x-pagination-total-pages'])

			if (!Array.isArray(response.data)) {
				return []
			}

			return response.data.map(entry => this.modelGetAllFactory(entry))
		} finally {
			cancel()
		}
	}

	/**
	 * Performs a put request to the url specified before
	 * @returns {Promise<any | never>}
	 */
	async create(model : Model) {
		if (this.paths.create === '') {
			throw new Error('This model is not able to create data.')
		}

		const cancel = this.setLoading()
		const finalUrl = this.getReplacedRoute(this.paths.create, model)

		try {
			const response = await this.http.put(finalUrl, model)
			const result = this.modelCreateFactory(response.data)
			if (typeof model.maxPermission !== 'undefined') {
				result.maxPermission = model.maxPermission
			}
			return result
		} finally {
			cancel()
		}
	}

	/**
	 * An abstract implementation to send post requests.
	 * Services can use this to implement functions to do post requests other than using the update method.
	 */
	async post(url : string, model : Model) {
		const cancel = this.setLoading()

		try {
			const response = await this.http.post(url, model)
			const result = this.modelUpdateFactory(response.data)
			if (typeof model.maxPermission !== 'undefined') {
				result.maxPermission = model.maxPermission
			}
			return result
		} finally {
			cancel()
		}
	}

	/**
	 * Performs a post request to the update url
	 */
	update(model : Model) {
		if (this.paths.update === '') {
			throw new Error('This model is not able to update data.')
		}

		const finalUrl = this.getReplacedRoute(this.paths.update, model)
		return this.post(finalUrl, model)
	}

	/**
	 * Performs a delete request to the update url
	 */
	async delete(model : Model) {
		if (this.paths.delete === '') {
			throw new Error('This model is not able to delete data.')
		}

		const cancel = this.setLoading()
		const finalUrl = this.getReplacedRoute(this.paths.delete, model)

		try {
			const {data} = await this.http.delete(finalUrl, model)
			return data
		} finally {
			cancel()
		}
	}

	/**
	 * Uploads a file to a url.
	 * @param url
	 * @param file {IFile}
	 * @param fieldName The name of the field the file is uploaded to.
	 */
	uploadFile(url : string, file: File, fieldName : string) {
		return this.uploadBlob(url, new Blob([file]), fieldName, file.name)
	}

	/**
	 * Uploads a blob to a url.
	 */
	uploadBlob(url : string, blob: Blob, fieldName: string, filename : string) {
		const data = new FormData()
		data.append(fieldName, blob, filename)
		return this.uploadFormData(url, data)
	}

	/**
	 * Uploads a form data object.
	 */
	async uploadFormData(url : string, formData: FormData) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.put(
				url,
				formData,
				{
					headers: {
						'Content-Type':
							'multipart/form-data; boundary=' + formData._boundary,
					},
					// fix upload issue after upgrading to axios to 1.0.0
					// see: https://github.com/axios/axios/issues/4885#issuecomment-1222419132
					transformRequest: formData => formData,
					onUploadProgress: ({progress}) => {
						this.uploadProgress = progress? Math.round((progress * 100)) : 0
					},
				},
			)
			return this.modelCreateFactory(response.data)
		} finally {
			this.uploadProgress = 0
			cancel()
		}
	}
}
