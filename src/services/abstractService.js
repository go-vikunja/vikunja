import axios from 'axios'
import {reduce, replace} from 'lodash'
import { objectToSnakeCase } from '../helpers/case'

export default class AbstractService {

	/////////////////////////////
	// Initial variable definitions
	///////////////////////////

	http = null
	loading = false
	paths = {
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
	 * @param paths An object with all paths. Default values are specified above.
	 */
	constructor(paths) {
		this.http = axios.create({
			baseURL: window.API_URL,
			headers: {
				'Content-Type': 'application/json',
			},
		})

		// Set the interceptors to process every request
		let self = this
		this.http.interceptors.request.use((config) => {
			switch (config.method) {
				case 'post':
					if (this.useUpdateInterceptor()) {
						config.data = self.beforeUpdate(config.data)
						config.data = JSON.stringify(objectToSnakeCase(config.data))
					}
					break
				case 'put':
					if (this.useCreateInterceptor()) {
						config.data = self.beforeCreate(config.data)
						config.data = JSON.stringify(objectToSnakeCase(config.data))
					}
					break
				case 'delete':
					if (this.useDeleteInterceptor()) {
						config.data = self.beforeDelete(config.data)
						config.data = JSON.stringify(objectToSnakeCase(config.data))
					}
					break
			}
			return config
		})

		// Set the default auth header if we have a token
		if (
			localStorage.getItem('token') !== '' &&
			localStorage.getItem('token') !== null &&
			localStorage.getItem('token') !== undefined
		) {
			this.http.defaults.headers.common['Authorization'] = 'Bearer ' + localStorage.getItem('token')
		}

		this.paths = {
			create: paths.create !== undefined ? paths.create : '',
			get: paths.get !== undefined ? paths.get : '',
			getAll: paths.getAll !== undefined ? paths.getAll : '',
			update: paths.update !== undefined ? paths.update : '',
			delete: paths.delete !== undefined ? paths.delete : '',
		}
	}

	/**
	 * Whether or not to use the create interceptor which processes a request payload into json
	 * @returns {boolean}
	 */
	useCreateInterceptor() {
		return true
	}

	/**
	 * Whether or not to use the update interceptor which processes a request payload into json
	 * @returns {boolean}
	 */
	useUpdateInterceptor() {
		return true
	}

	/**
	 * Whether or not to use the delete interceptor which processes a request payload into json
	 * @returns {boolean}
	 */
	useDeleteInterceptor() {
		return true
	}

	/////////////////////
	// Global error handler
	///////////////////

	/**
	 * Handles the error and rejects the promise.
	 * @param error
	 * @returns {Promise<never>}
	 */
	errorHandler(error) {
		return Promise.reject(error)
	}

	/////////////////
	// Helper functions
	///////////////

	/**
	 * Returns an object with all route parameters and their values.
	 * @param route
	 * @returns object
	 */
	getRouteReplacements(route) {
		let parameters = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {}
		let replace$$1 = {}
		let pattern = this.getRouteParameterPattern()
		pattern = new RegExp(pattern instanceof RegExp ? pattern.source : pattern, 'g')

		for (let parameter; (parameter = pattern.exec(route)) !== null;) {
			replace$$1[parameter[0]] = parameters[parameter[1]];
		}

		return replace$$1;
	}

	/**
	 * Holds the replacement pattern for url paths, can be overwritten by implementations.
	 * @return {RegExp}
	 */
	getRouteParameterPattern() {
		return /{([^}]+)}/
	}

	/**
	 * Returns a fully-ready-ready-to-make-a-request-to route with replaced parameters.
	 * @param path
	 * @param pathparams
	 * @return string
	 */
	getReplacedRoute(path, pathparams) {
		let replacements = this.getRouteReplacements(path, pathparams)
		return reduce(replacements, function (result, value, parameter) {
			return replace(result, parameter, value)
		}, path)
	}

	/**
	 * setLoading is a method which sets the loading variable to true, after a timeout of 100ms.
	 * It has the timeout to prevent the loading indicator from showing for only a blink of an eye in the
	 * case the api returns a response in < 100ms.
	 * But because the timeout is created using setTimeout, it will still trigger even if the request is
	 * already finished, so we return a method to call in that case.
	 * @returns {Function}
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
	// Specific factories for each request are completly optional, if these are not specified, the defautl factory is used.
	////////////////

	/**
	 * The modelFactory returns an model from an object.
	 * This one here is the default one, usually the service definitions for a model will override this.
	 * @param data
	 * @returns {*}
	 */
	modelFactory(data) {
		return data
	}

	/**
	 * This is the model factory for get requests.
	 * @param data
	 * @return {*}
	 */
	modelGetFactory(data) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for get all requests.
	 * @param data
	 * @return {*}
	 */
	modelGetAllFactory(data) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for create requests.
	 * @param data
	 * @return {*}
	 */
	modelCreateFactory(data) {
		return this.modelFactory(data)
	}

	/**
	 * This is the model factory for update requests.
	 * @param data
	 * @return {*}
	 */
	modelUpdateFactory(data) {
		return this.modelFactory(data)
	}

	//////////////
	// Preprocessors
	////////////

	/**
	 * Default preprocessor for get requests
	 * @param model
	 * @return {*}
	 */
	beforeGet(model) {
		return model
	}

	/**
	 * Default preprocessor for create requests
	 * @param model
	 * @return {*}
	 */
	beforeCreate(model) {
		return model
	}

	/**
	 * Default preprocessor for update requests
	 * @param model
	 * @return {*}
	 */
	beforeUpdate(model) {
		return model
	}

	/**
	 * Default preprocessor for delete requests
	 * @param model
	 * @return {*}
	 */
	beforeDelete(model) {
		return model
	}

	///////////////
	// Global actions
	/////////////

	/**
	 * Performs a get request to the url specified before.
	 * @param model The model to use. The request path is built using the values from the model.
	 * @param params Optional query parameters
	 * @returns {Q.Promise<any>}
	 */
	get(model, params = {}) {
		if (this.paths.get === '') {
			return Promise.reject({message: 'This model is not able to get data.'})
		}

		return this.getM(this.paths.get, model, params)
	}

	/**
	 * This is a more abstract implementation which only does a get request.
	 * Services which need more flexibility can use this.
	 * @param url
	 * @param model
	 * @param params
	 * @returns {Q.Promise<unknown>}
	 */
	getM(url, model = {}, params = {}) {
		const cancel = this.setLoading()

		model = this.beforeGet(model)
		const finalUrl = this.getReplacedRoute(url, model)

		return this.http.get(finalUrl, {params: params})
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				return Promise.resolve(this.modelGetFactory(response.data))
			})
			.finally(() => {
				cancel()
			})
	}

	/**
	 * Performs a get request to the url specified before.
	 * The difference between this and get() is this one is used to get a bunch of data (an array), not just a single object.
	 * @param model The model to use. The request path is built using the values from the model.
	 * @param params Optional query parameters
	 * @param page The page to get
	 * @returns {Q.Promise<any>}
	 */
	getAll(model = {}, params = {}, page = 1) {
		if (this.paths.getAll === '') {
			return Promise.reject({message: 'This model is not able to get data.'})
		}

		params.page = page

		const cancel = this.setLoading()
		model = this.beforeGet(model)
		const finalUrl = this.getReplacedRoute(this.paths.getAll, model)

		return this.http.get(finalUrl, {params: params})
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				this.resultCount = Number(response.headers['x-pagination-result-count'])
				this.totalPages = Number(response.headers['x-pagination-total-pages'])

				if (Array.isArray(response.data)) {
					return Promise.resolve(response.data.map(entry => {
						return this.modelGetAllFactory(entry)
					}))
				}
				if (response.data === null) {
					return Promise.resolve([])
				}
				return Promise.resolve(this.modelGetAllFactory(response.data))
			})
			.finally(() => {
				cancel()
			})
	}

	/**
	 * Performs a put request to the url specified before
	 * @param model
	 * @returns {Promise<any | never>}
	 */
	create(model) {
		if (this.paths.create === '') {
			return Promise.reject({message: 'This model is not able to create data.'})
		}

		const cancel = this.setLoading()
		const finalUrl = this.getReplacedRoute(this.paths.create, model)

		return this.http.put(finalUrl, model)
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				return Promise.resolve(this.modelCreateFactory(response.data))
			})
			.finally(() => {
				cancel()
			})
	}

	/**
	 * An abstract implementation to send post requests.
	 * Services can use this to implement functions to do post requests other than using the update method.
	 * @param url
	 * @param model
	 * @returns {Q.Promise<unknown>}
	 */
	post(url, model) {
		const cancel = this.setLoading()

		return this.http.post(url, model)
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				return Promise.resolve(this.modelUpdateFactory(response.data))
			})
			.finally(() => {
				cancel()
			})
	}

	/**
	 * Performs a post request to the update url
	 * @param model
	 * @returns {Q.Promise<any>}
	 */
	update(model) {
		if (this.paths.update === '') {
			return Promise.reject({message: 'This model is not able to update data.'})
		}

		const finalUrl = this.getReplacedRoute(this.paths.update, model)
		return this.post(finalUrl, model)
	}

	/**
	 * Performs a delete request to the update url
	 * @param model
	 * @returns {Q.Promise<any>}
	 */
	delete(model) {
		if (this.paths.delete === '') {
			return Promise.reject({message: 'This model is not able to delete data.'})
		}

		const cancel = this.setLoading()
		const finalUrl = this.getReplacedRoute(this.paths.delete, model)

		return this.http.delete(finalUrl, model)
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				return Promise.resolve(response.data)
			})
			.finally(() => {
				cancel()
			})
	}
}