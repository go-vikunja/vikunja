import { AsyncLocalStorage } from 'node:async_hooks';

/**
 * Request context that flows through async operations
 */
export interface RequestContext {
	sessionId: string;
	userId?: number;
	username?: string;
}

/**
 * AsyncLocalStorage for tracking request context across async operations
 * This allows us to correlate MCP protocol handlers with the HTTP session
 * that initiated the request, without needing to modify the MCP SDK's
 * request handler signatures.
 */
export const requestContext = new AsyncLocalStorage<RequestContext>();

/**
 * Get the current request context
 * Returns undefined if not within a request context
 */
export function getRequestContext(): RequestContext | undefined {
	return requestContext.getStore();
}

/**
 * Get the current session ID from request context
 * Throws if not within a request context
 */
export function getCurrentSessionId(): string {
	const context = getRequestContext();
	if (!context) {
		throw new Error('No request context available');
	}
	return context.sessionId;
}

/**
 * Run a function within a request context
 */
export function runInContext<T>(context: RequestContext, fn: () => T | Promise<T>): T | Promise<T> {
	return requestContext.run(context, fn);
}
