import {describe, it, expect} from 'vitest'
import {AxiosError} from 'axios'

import {shouldDropEvent} from './sentryFilters'

// Object.assign instead of `new Error(msg, {cause})`: the vitest tsconfig
// targets a lib without the two-argument Error constructor.
function errorWithCause(message: string, cause: unknown): Error {
	return Object.assign(new Error(message), {cause})
}

describe('shouldDropEvent', () => {
	it('drops a plain AxiosError', () => {
		expect(shouldDropEvent(new AxiosError('Request failed'))).toBe(true)
	})

	it('drops an error wrapping an AxiosError as cause', () => {
		expect(shouldDropEvent(errorWithCause('Error renewing token: ', new AxiosError('Request failed')))).toBe(true)
	})

	it('drops an error with an AxiosError two levels deep', () => {
		const inner = errorWithCause('inner', new AxiosError('Request failed'))

		expect(shouldDropEvent(errorWithCause('outer', inner))).toBe(true)
	})

	it('drops an error-like object with code and message', () => {
		expect(shouldDropEvent({code: 'ECONNABORTED', message: 'timeout'})).toBe(true)
	})

	it('keeps a plain error', () => {
		expect(shouldDropEvent(new Error('something actually broke'))).toBe(false)
	})

	it('keeps a plain error wrapping another plain error', () => {
		expect(shouldDropEvent(errorWithCause('outer', new Error('inner')))).toBe(false)
	})

	it('keeps undefined', () => {
		expect(shouldDropEvent(undefined)).toBe(false)
	})

	it('does not loop on a cause cycle', () => {
		const a = new Error('a')
		const b = errorWithCause('b', a)
		Object.assign(a, {cause: b})

		expect(shouldDropEvent(a)).toBe(false)
	})
})

describe('shouldDropEvent with chunk load errors', () => {
	const messages = [
		'Failed to fetch dynamically imported module: https://try.vikunja.io/assets/ProjectList-abc123.js',
		'error loading dynamically imported module: https://try.vikunja.io/assets/ProjectList-abc123.js',
		'Importing a module script failed.',
		'Unable to preload CSS for /assets/ProjectList-abc123.css',
		'\'text/html\' is not a valid JavaScript MIME type.',
		'Loading module from “https://try.vikunja.io/assets/ProjectList-abc123.js” was blocked because of a disallowed MIME type (“text/html”).',
		'Failed to load module script: Expected a JavaScript module script but the server responded with a MIME type of "text/html".',
	]

	it.each(messages)('drops the exception %s', message => {
		expect(shouldDropEvent(new Error(message))).toBe(true)
	})

	it.each(messages)('drops the event message %s', message => {
		expect(shouldDropEvent(undefined, {message})).toBe(true)
	})

	it.each(messages)('drops the event exception value %s', message => {
		expect(shouldDropEvent(undefined, {exception: {values: [{value: message}]}})).toBe(true)
	})

	it('keeps an unrelated event message', () => {
		expect(shouldDropEvent(undefined, {message: 'something actually broke'})).toBe(false)
	})

	it('keeps an unrelated mime type error', () => {
		expect(shouldDropEvent(new Error('Refused to apply style because its MIME type is not supported'))).toBe(false)
	})
})

describe('shouldDropEvent with stale chunk fallout', () => {
	const messages = [
		'Couldn\'t resolve component "default" at "/tasks/:id"',
		'Couldn\'t resolve component "default" at "/projects"',
		'Couldn\'t resolve component "default" at "/projects". Ensure you passed a function that returns a promise.',
		'Async component timed out after 60000ms.',
	]

	it.each(messages)('drops the exception %s', message => {
		expect(shouldDropEvent(new Error(message))).toBe(true)
	})

	it.each(messages)('drops the event exception value %s', message => {
		expect(shouldDropEvent(undefined, {exception: {values: [{value: message}]}})).toBe(true)
	})

	it('keeps an unrelated component error', () => {
		expect(shouldDropEvent(new Error('Failed to mount component: template or render function not defined'))).toBe(false)
	})

	it('keeps an unrelated timeout', () => {
		expect(shouldDropEvent(new Error('Navigation timed out after 5000ms'))).toBe(false)
	})
})

describe('shouldDropEvent with third party injections', () => {
	const messages = [
		'Invalid call to runtime.sendMessage(). Tab not found.',
		'undefined is not an object (evaluating \'window.webkit.messageHandlers\')',
		'Error invoking postMessage: Java object is gone',
		'undefined is not an object (evaluating \'window.weixinPostMessageHandlers.weixinDispatchMessage.postMessage\')',
		'WKWebView API client did not respond to this postMessage',
	]

	it.each(messages)('drops the exception %s', message => {
		expect(shouldDropEvent(new Error(message))).toBe(true)
	})

	it.each(messages)('drops the event exception value %s', message => {
		expect(shouldDropEvent(undefined, {exception: {values: [{value: message}]}})).toBe(true)
	})

	const extensionUrls = [
		'chrome-extension://abcdefghijklmnop/content.js',
		'moz-extension://abcdefghijklmnop/content.js',
		'safari-web-extension://ABCDEF-1234/content.js',
		'iabjs://navigation_performance_logger_android',
	]

	it.each(extensionUrls)('drops an event thrown from %s', filename => {
		expect(shouldDropEvent(undefined, {
			exception: {
				values: [{
					value: 'boom',
					stacktrace: {frames: [{filename: 'https://try.vikunja.io/assets/index.js'}, {filename}]},
				}],
			},
		})).toBe(true)
	})

	it('keeps an event an extension only appears deeper in', () => {
		expect(shouldDropEvent(undefined, {
			exception: {
				values: [{
					value: 'boom',
					stacktrace: {frames: [{filename: 'chrome-extension://abc/content.js'}, {filename: 'https://try.vikunja.io/assets/index.js'}]},
				}],
			},
		})).toBe(false)
	})

	it('keeps an event thrown from our own code', () => {
		expect(shouldDropEvent(undefined, {
			exception: {
				values: [{
					value: 'boom',
					stacktrace: {frames: [{filename: 'https://try.vikunja.io/assets/index.js'}]},
				}],
			},
		})).toBe(false)
	})

	it('keeps an unrelated client error', () => {
		expect(shouldDropEvent(new Error('The API client did not respond in time'))).toBe(false)
	})

	it('keeps an unrelated postMessage error', () => {
		expect(shouldDropEvent(new Error('Failed to execute \'postMessage\' on \'Window\''))).toBe(false)
	})
})

describe('shouldDropEvent with empty events', () => {
	it('drops an event without a message or exception', () => {
		expect(shouldDropEvent(undefined, {})).toBe(true)
	})

	it('drops an event whose exception values are blank', () => {
		expect(shouldDropEvent(undefined, {exception: {values: [{}]}})).toBe(true)
	})

	it('keeps an event with only an exception type', () => {
		expect(shouldDropEvent(undefined, {exception: {values: [{type: 'TypeError'}]}})).toBe(false)
	})

	it('keeps an event with only a message', () => {
		expect(shouldDropEvent(undefined, {message: 'something actually broke'})).toBe(false)
	})

	it('keeps an event when no event was passed at all', () => {
		expect(shouldDropEvent(new Error('something actually broke'))).toBe(false)
	})
})
