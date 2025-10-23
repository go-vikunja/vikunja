import type { Request, Response } from 'express';
import { SSEServerTransport } from '@modelcontextprotocol/sdk/server/sse.js';
import type { VikunjaMCPServer } from '../../server.js';
import type { SessionManager } from './session-manager.js';
import type { TokenValidator } from '../../auth/token-validator.js';
import type { RateLimiter } from '../../ratelimit/limiter.js';
import { logHttpTransport, logAuth, logError, logger } from '../../utils/logger.js';
import { validateSSEMessage } from '../../utils/request-validation.js';

/**
 * SSE Transport configuration
 */
export interface SSETransportConfig {
	mcpServer: VikunjaMCPServer;
	sessionManager: SessionManager;
	tokenValidator: TokenValidator;
	rateLimiter: RateLimiter;
	postEndpoint?: string; // Default: '/sse' for POST messages
}

/**
 * Express middleware-compatible wrapper for MCP SSEServerTransport
 * 
 * ⚠️ DEPRECATED: This transport is deprecated in favor of HTTP Streamable.
 * Use HTTP Streamable for new integrations. SSE support will be removed in v2.0.
 * 
 * Why Deprecated:
 * - EventSource API limitations (no custom headers)
 * - Requires query parameter authentication (less secure)
 * - Two-endpoint complexity (GET for events, POST for messages)
 * - Being phased out by MCP client tools (n8n, etc.)
 * 
 * This class integrates the MCP SDK's SSE transport with our
 * authentication, rate limiting, and session management infrastructure.
 * 
 * Flow:
 * 1. Client opens GET /sse with token in query param
 * 2. Server validates token, creates session, establishes SSE stream
 * 3. Server sends session_id in initial SSE event
 * 4. Client uses session_id for POST /sse requests
 * 5. Server routes POST messages to correct MCP server instance
 * 
 * Usage:
 * ```typescript
 * const transport = new SSETransport({ 
 *   mcpServer, sessionManager, tokenValidator, rateLimiter 
 * });
 * app.get('/sse', transport.handleGetStream.bind(transport));
 * app.post('/sse', transport.handlePostMessage.bind(transport));
 * ```
 */
export class SSETransport {
	private readonly mcpServer: VikunjaMCPServer;
	private readonly sessionManager: SessionManager;
	private readonly tokenValidator: TokenValidator;
	private readonly rateLimiter: RateLimiter;
	private readonly postEndpoint: string;
	private readonly transports = new Map<string, SSEServerTransport>();

	constructor(config: SSETransportConfig) {
		this.mcpServer = config.mcpServer;
		this.sessionManager = config.sessionManager;
		this.tokenValidator = config.tokenValidator;
		this.rateLimiter = config.rateLimiter;
		this.postEndpoint = config.postEndpoint || '/sse';

		// Log deprecation warning on instantiation
		logger.warn('SSE transport initialized (DEPRECATED)', {
			message: 'SSE transport is deprecated and will be removed in v2.0',
			migration: 'Use HTTP Streamable transport instead',
			documentation: 'See docs/migration-sse-to-http-streamable.md',
		});
	}

	/**
	 * Extract token from query parameter or Authorization header
	 * 
	 * EventSource API limitation: Cannot send custom headers, so token must be in query param for browsers.
	 * However, we support Authorization header for non-browser clients (curl, postman, etc.)
	 */
	private extractQueryToken(req: Request): string | null {
		// 1. Try query parameter first (required for EventSource API)
		const queryToken = req.query['token'];
		if (typeof queryToken === 'string' && queryToken.length > 0) {
			return queryToken;
		}

		// 2. Fallback to Authorization header for non-browser clients
		const authHeader = req.headers.authorization;
		if (authHeader && typeof authHeader === 'string') {
			const match = authHeader.match(/^Bearer\s+(.+)$/i);
			if (match && match[1]) {
				return match[1];
			}
		}

		return null;
	}	/**
	 * Express middleware handler for GET /sse requests (event stream establishment)
	 * 
	 * Flow:
	 * 1. Extract and validate token from query parameter
	 * 2. Check rate limits
	 * 3. Create session
	 * 4. Create SSEServerTransport and connect to MCP server
	 * 5. Send deprecation warning in SSE event
	 * 6. Stream MCP messages as SSE events
	 * 
	 * Error handling:
	 * - 401: Missing or invalid authentication token
	 * - 429: Rate limit exceeded
	 * - 500: Internal server error (session creation, transport setup, etc.)
	 */
	async handleStream(req: Request, res: Response): Promise<void> {
		const startTime = Date.now();

		try {
			// 1. Extract token from query parameter
			const token = this.extractQueryToken(req);
			if (!token) {
				logAuth('authentication_failed', undefined, {
					reason: 'missing_query_token',
					path: req.path,
					ip: req.ip,
					userAgent: req.headers['user-agent'],
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication required: Token query parameter missing',
						data: {
							reason: 'EventSource API requires ?token=xxx in URL',
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 2. Authenticate token
			let userContext;
			try {
				userContext = await this.tokenValidator.validateToken(token);
			} catch (error) {
				logAuth('authentication_failed', undefined, {
					reason: 'invalid_token',
					error: error instanceof Error ? error.message : String(error),
					path: req.path,
					ip: req.ip,
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication failed: Invalid or expired token',
						data: {
							reason: error instanceof Error ? error.message : 'Token validation failed',
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 3. Check rate limits
			try {
				await this.rateLimiter.checkLimit(token);
			} catch (error) {
				const retryAfter = (error as any).retryAfter || 60;
				logAuth('auth_failed', undefined, {
					reason: 'rate_limit_exceeded',
					token: token.substring(0, 10) + '...',
					path: req.path,
				});
				res.set('Retry-After', String(retryAfter));
				res.status(429).json({
					error: {
						code: -32002,
						message: 'Rate limit exceeded',
						data: {
							retryAfter,
							limit: 100,
							window: 900,
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 4. Create session
			let session;
			try {
				session = this.sessionManager.createSession(
					token,
					userContext,
					'sse'
				);
			} catch (error) {
				logError(error instanceof Error ? error : new Error(String(error)), {
					userId: userContext.userId,
					transport: 'sse',
					context: 'Session creation failed',
				});
				res.status(500).json({
					error: {
						code: -32603,
						message: 'Internal server error: Session creation failed',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			logAuth('token_validated', undefined, {
				sessionId: session.id,
				transport: 'sse',
				ip: req.ip,
				userId: userContext.userId,
			});

			// 5. Add deprecation warning header (must be done BEFORE transport.start())
			res.setHeader('X-Deprecation', 'SSE transport is deprecated. Migrate to HTTP Streamable (POST /mcp).');

			// 6. Create SSEServerTransport
			const transport = new SSEServerTransport(this.postEndpoint, res as any);
			
			// T039: Use SDK's transport.sessionId everywhere for consistency
			// The SDK generates its own session ID which we must respect
			const transportSessionId = transport.sessionId;
			
			// Store transport using SDK's session ID (this is what POST requests will use)
			this.transports.set(transportSessionId, transport);
			
			// Also update our SessionManager to use the SDK's ID
			// This ensures session.id matches transport.sessionId
			(session as any).id = transportSessionId;

			// Cleanup on connection close
			res.on('close', () => {
				this.transports.delete(transportSessionId);
				this.sessionManager.terminateSession(transportSessionId);
				logHttpTransport('disconnect', {
					sessionId: transportSessionId,
					userId: userContext.userId,
					duration: Date.now() - startTime,
				});
			});

			// Set up transport event handlers
			transport.onclose = () => {
				this.transports.delete(transportSessionId);
				this.sessionManager.terminateSession(transportSessionId);
				logHttpTransport('disconnect', {
					sessionId: transportSessionId,
					userId: userContext.userId,
				});
			};

			transport.onerror = (error: Error) => {
				logError(error, {
					context: 'sse_transport_error',
					sessionId: transportSessionId,
					userId: userContext.userId,
				});
			};

			// 7. Connect transport to MCP server
			try {
				// T039/T040: connect() automatically calls start(), which sends initial events
				// So we DON'T call transport.start() separately
				await this.mcpServer.getServer().connect(transport);
				// Set user context using the SDK's session ID
				this.mcpServer.setUserContext(transportSessionId, userContext);
				
				// T040: Send custom session event AFTER SDK initialization
				// The SDK's start() (called by connect()) sets up headers and begins streaming
				// Now we send our custom session event with deprecation info
				const sessionEventData = JSON.stringify({
					session_id: transportSessionId,
					deprecated: true,
					deprecation_message: 'SSE transport is deprecated. Migrate to HTTP Streamable (POST /mcp).',
					migration: 'See docs/migration-sse-to-http-streamable.md',
				});
				res.write(`event: session\n`);
				res.write(`data: ${sessionEventData}\n\n`);
			} catch (error) {
				logError(error as Error, {
					context: 'mcp_server_connection_failed',
					sessionId: transportSessionId,
					userId: userContext.userId,
				});
				this.transports.delete(transportSessionId);
				this.sessionManager.terminateSession(transportSessionId);
				// At this point, transport.start() may have already been called,
				// so we can't send JSON response. The transport will handle errors.
				return;
			}

			logHttpTransport('connection', {
				sessionId: transportSessionId,
				userId: userContext.userId,
				duration: Date.now() - startTime,
				transport: 'sse',
			});

			// Note: Connection stays open for SSE stream
		} catch (error) {
			logError(error as Error, {
				context: 'sse_get_handler_error',
				path: req.path,
				ip: req.ip,
			});

			// Only send response if headers haven't been sent yet
			if (!res.headersSent) {
				res.status(500).json({
					error: {
						code: -32603,
						message: 'Internal server error',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
			}
		}
	}

	/**
	 * Express middleware handler for POST /sse requests (client message routing)
	 * 
	 * Flow:
	 * 1. Extract and validate token from query parameter
	 * 2. Extract session_id and message from request body
	 * 3. Validate session exists
	 * 4. Check rate limits
	 * 5. Route message to SSEServerTransport
	 * 6. Response is sent via SSE stream (202 Accepted)
	 * 
	 * Error handling:
	 * - 400: Missing session_id or message in body
	 * - 401: Missing or invalid authentication token
	 * - 404: Session not found or expired
	 * - 429: Rate limit exceeded
	 * - 500: Internal server error
	 */
	async handleMessage(req: Request, res: Response): Promise<void> {
		const startTime = Date.now();

		try {
			// 1. Extract token from query parameter
			const token = this.extractQueryToken(req);
			if (!token) {
				logAuth('authentication_failed', undefined, {
					reason: 'missing_query_token',
					path: req.path,
					ip: req.ip,
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication required: Token query parameter missing',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 2. Validate request body structure with Zod schema
			// Protects against malformed JSON, oversized payloads, and injection attacks
			const validationResult = validateSSEMessage(req.body);
			if (!validationResult.success) {
				res.status(400).json({
					...validationResult.error,
					error: {
						...validationResult.error.error,
						data: {
							...validationResult.error.error.data,
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				logError(new Error('SSE message validation failed'), {
					event: 'sse_validation_failed',
					token: token ? token.substring(0, 8) : undefined,
				});
				return;
			}

			const { session_id, message } = validationResult.data;

			// 3. Validate session exists
			const session = this.sessionManager.getSession(session_id);
			if (!session) {
				res.status(404).json({
					error: {
						code: -32003,
						message: 'Session not found or expired',
						data: {
							session_id,
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 4. Validate token matches session
			if (session.token !== token) {
				logAuth('auth_failed', undefined, {
					reason: 'token_mismatch',
					sessionId: session_id,
					userId: String(session.userContext.userId),
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication failed: Token does not match session',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 5. Check rate limits
			try {
				await this.rateLimiter.checkLimit(token);
			} catch (error) {
				const retryAfter = (error as any).retryAfter || 60;
				logAuth('auth_failed', undefined, {
					reason: 'rate_limit_exceeded',
					sessionId: session_id,
					token: token.substring(0, 10) + '...',
					userId: String(session.userContext.userId),
				});
				res.set('Retry-After', String(retryAfter));
				res.status(429).json({
					error: {
						code: -32002,
						message: 'Rate limit exceeded',
						data: {
							retryAfter,
							limit: 100,
							window: 900,
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 6. Get transport for session
			const transport = this.transports.get(session_id);
			if (!transport) {
				logError(new Error('Missing transport'), {
					context: 'transport_not_found_for_session',
					sessionId: session_id,
					userId: session.userContext.userId,
				});
				res.status(500).json({
					error: {
						code: -32603,
						message: 'Internal server error: Transport not found for session',
						data: {
							hint: 'SSE stream may have been closed. Re-establish GET /sse connection.',
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// 7. Update session activity
			this.sessionManager.updateActivity(session_id);

			// 8. T041: Route message to SDK transport for processing
			// The transport will process the message and send response via SSE stream (GET connection)
			try {
				await transport.handleMessage(message);
				
				// Return 202 Accepted immediately (response goes via SSE stream)
				res.status(202).json({ accepted: true });

				logHttpTransport('request_handled', {
					sessionId: session_id,
					userId: session.userContext.userId,
					messageId: (message as any).id,
					method: (message as any).method,
					duration: Date.now() - startTime,
				});
			} catch (messageError) {
				logError(messageError as Error, {
					context: 'sse_message_routing_error',
					sessionId: session_id,
					userId: session.userContext.userId,
				});
				res.status(500).json({
					error: {
						code: -32603,
						message: 'Internal server error: Message routing failed',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
				return;
			}

			// Note: Response has been sent (202 Accepted)
		} catch (error) {
			logError(error as Error, {
				context: 'sse_post_handler_error',
				path: req.path,
				ip: req.ip,
			});

			if (!res.headersSent) {
				res.status(500).json({
					error: {
						code: -32603,
						message: 'Internal server error',
						data: {
							deprecation: 'SSE transport is deprecated. Use HTTP Streamable instead.',
						},
					},
				});
			}
		}
	}

	/**
	 * Get active SSE transport by session ID
	 * 
	 * Used for debugging and monitoring.
	 */
	getTransport(sessionId: string): SSEServerTransport | undefined {
		return this.transports.get(sessionId);
	}

	/**
	 * Get count of active SSE connections
	 */
	getActiveConnectionCount(): number {
		return this.transports.size;
	}

	/**
	 * Cleanup: Close all active SSE transports
	 * 
	 * Should be called on server shutdown.
	 */
	async closeAll(): Promise<void> {
		logger.warn('Closing all SSE transports', {
			count: this.transports.size,
		});

		const closePromises: Promise<void>[] = [];
		for (const [sessionId, transport] of this.transports.entries()) {
			closePromises.push(
				transport.close().catch((error) => {
					logError(error as Error, {
						context: 'error_closing_sse_transport',
						sessionId,
					});
				})
			);
		}

		await Promise.all(closePromises);
		this.transports.clear();

		logger.warn('All SSE transports closed', {
			count: closePromises.length,
		});
	}
}
