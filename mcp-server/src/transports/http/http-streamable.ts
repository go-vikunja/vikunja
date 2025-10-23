import type { Request, Response } from 'express';
import { StreamableHTTPServerTransport } from '@modelcontextprotocol/sdk/server/streamableHttp.js';
import type { IncomingMessage, ServerResponse } from 'node:http';
import type { VikunjaMCPServer } from '../../server.js';
import type { SessionManager } from './session-manager.js';
import type { TokenValidator } from '../../auth/token-validator.js';
import type { RateLimiter } from '../../ratelimit/limiter.js';
import { logHttpTransport, logAuth, logError } from '../../utils/logger.js';
import { runInContext } from '../../utils/request-context.js';
import { validateJsonRpcRequest } from '../../utils/request-validation.js';

/**
 * HTTP Streamable Transport configuration
 */
export interface HTTPStreamableTransportConfig {
	mcpServer: VikunjaMCPServer;
	sessionManager: SessionManager;
	tokenValidator: TokenValidator;
	rateLimiter: RateLimiter;
	enableJsonResponse?: boolean;
}

/**
 * Express middleware-compatible wrapper for MCP StreamableHTTPServerTransport
 * 
 * Implements shared VikunjaMCPServer architecture: one MCP server instance with
 * session-specific user contexts passed through per-request.
 * 
 * This class integrates the MCP SDK's HTTP Streamable transport with our
 * authentication, rate limiting, and session management infrastructure.
 * 
 * Usage:
 * ```typescript
 * const transport = new HTTPStreamableTransport({ 
 *   mcpServer, sessionManager, tokenValidator, rateLimiter 
 * });
 * app.post('/mcp', transport.handleRequest.bind(transport));
 * ```
 */
export class HTTPStreamableTransport {
	private readonly mcpServer: VikunjaMCPServer;
	private readonly sessionManager: SessionManager;
	private readonly tokenValidator: TokenValidator;
	private readonly rateLimiter: RateLimiter;
	private readonly enableJsonResponse: boolean;
	private readonly transports = new Map<string, StreamableHTTPServerTransport>();

	constructor(config: HTTPStreamableTransportConfig) {
		this.mcpServer = config.mcpServer;
		this.sessionManager = config.sessionManager;
		this.tokenValidator = config.tokenValidator;
		this.rateLimiter = config.rateLimiter;
		this.enableJsonResponse = config.enableJsonResponse ?? false;
	}

	/**
	 * Extract Bearer token from Authorization header
	 */
	private extractBearerToken(req: Request): string | null {
		const authHeader = req.headers.authorization;
		if (!authHeader) {
			return null;
		}

		const parts = authHeader.split(' ');
		if (parts.length !== 2 || parts[0] !== 'Bearer') {
			return null;
		}

		return parts[1] || null;
	}

	/**
	 * Express middleware handler for POST /mcp requests
	 * 
	 * Flow:
	 * 1. Extract and validate Bearer token
	 * 2. Check rate limits
	 * 3. Create or retrieve session
	 * 4. Delegate to StreamableHTTPServerTransport
	 * 
	 * Error handling:
	 * - 401: Missing or invalid authentication token
	 * - 429: Rate limit exceeded
	 * - 500: Internal server error (session creation, transport setup, etc.)
	 */
	async handleRequest(req: Request, res: Response): Promise<void> {
		const startTime = Date.now();
		let sessionId: string | undefined;
		
		try {
			// 0. If JSON response mode is enabled, inject Accept header to bypass SDK validation
			// The MCP SDK requires Accept: application/json, text/event-stream, but n8n and other
			// clients can't customize this header. When enableJsonResponse is true, we inject it.
			if (this.enableJsonResponse && (!req.headers.accept || !req.headers.accept.includes('text/event-stream'))) {
				req.headers.accept = 'application/json, text/event-stream';
			}

			// 1. Extract token
			const token = this.extractBearerToken(req);
			if (!token) {
				logAuth('authentication_failed', undefined, {
					reason: 'missing_token',
					path: req.path,
					ip: req.ip,
					userAgent: req.headers['user-agent'],
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication required: Bearer token missing',
					},
				});
				return;
			}

			// 2. Validate token and get user context
			let userContext;
			try {
				userContext = await this.tokenValidator.validateToken(token);
				logAuth('token_validated', token.substring(0, 8), {
					userId: userContext.userId,
					username: userContext.username,
					cached: false, // TODO: Track if from cache
				});
			} catch (error) {
				logAuth('authentication_failed', token.substring(0, 8), {
					reason: 'invalid_token',
					error: error instanceof Error ? error.message : 'Unknown error',
					ip: req.ip,
				});
				res.status(401).json({
					error: {
						code: -32001,
						message: 'Authentication failed: Invalid token',
					},
				});
				return;
			}

			// 3. Check rate limits
			try {
				await this.rateLimiter.checkLimit(token);
			} catch (error) {
				logHttpTransport('rate_limit_exceeded', {
					token: token.substring(0, 8),
					userId: userContext.userId,
					error: error instanceof Error ? error.message : 'Unknown error',
				});
				res.status(429).json({
					error: {
						code: -32003,
						message: 'Rate limit exceeded',
						data: {
							retryAfter: 60, // Default retry after 60 seconds
						},
					},
				});
				return;
			}

			// 4. Get or create session
			// Note: SDK uses 'mcp-session-id' header (lowercase in Express)
			sessionId = req.headers['mcp-session-id'] as string | undefined;
			let session;

			if (sessionId) {
				session = this.sessionManager.getSession(sessionId);
				if (session) {
					this.sessionManager.updateActivity(sessionId);
					logHttpTransport('session_activity_updated', {
						sessionId,
						userId: userContext.userId,
					});
				} else {
					// Session expired or invalid, create new one
					logHttpTransport('session_expired', {
						sessionId,
						userId: userContext.userId,
						reason: 'session_not_found',
					});
					sessionId = undefined;
				}
			}

			if (!session) {
				session = this.sessionManager.createSession(token, userContext, 'http-streamable');
				sessionId = session.id;
				logHttpTransport('session_created', {
					sessionId,
					userId: userContext.userId,
					username: userContext.username,
					transport: 'http-streamable',
					ip: req.ip,
					userAgent: req.headers['user-agent'],
				});
			}

			// 5. Get or create StreamableHTTPServerTransport for this session
			// At this point, sessionId is guaranteed to be defined
			if (!sessionId) {
				throw new Error('Session ID not set');
			}

			let transport = this.transports.get(sessionId);
			if (!transport) {
				logHttpTransport('transport_creating', {
					sessionId,
					userId: userContext.userId,
					enableJsonResponse: this.enableJsonResponse,
				});
				
				transport = new StreamableHTTPServerTransport({
					sessionIdGenerator: () => sessionId!,
					enableJsonResponse: this.enableJsonResponse,
					onsessioninitialized: async (sid: string) => {
						logHttpTransport('mcp_session_initialized', { sessionId: sid });
						// Store user context in shared MCP server for this session
						// Add token to userContext for auth/types.UserContext compatibility
						this.mcpServer.setUserContext(sid, { ...userContext, token, email: userContext.email || '' });
					},
					onsessionclosed: async (sid: string) => {
						logHttpTransport('mcp_session_closed', { sessionId: sid });
						// Remove user context from shared MCP server
						this.mcpServer.removeUserContext(sid);
						// Clean up transport and session
						this.transports.delete(sid);
						this.sessionManager.terminateSession(sid);
					},
				});

				// Connect this transport to the shared MCP server
				// The Server instance will handle all MCP protocol messages via this transport
				try {
					await this.mcpServer.getServer().connect(transport);
					logHttpTransport('transport_connected', {
						sessionId,
						userId: userContext.userId,
					});
				} catch (error) {
					logError(error as Error, {
						context: 'transport_connect_failed',
						sessionId,
						userId: userContext.userId,
					});
					throw new Error(`Failed to connect transport: ${error instanceof Error ? error.message : 'Unknown error'}`);
				}

				this.transports.set(sessionId, transport);
			} else {
				// Existing transport - update user context in case token was refreshed
				this.mcpServer.setUserContext(sessionId, { ...userContext, token, email: userContext.email || '' });
			}

			// Note: SDK automatically manages 'Mcp-Session-Id' header - no need to set it manually

			// 7. Validate request body before SDK processing
			// Protects against malformed JSON, oversized payloads, and injection attacks
			const validationResult = validateJsonRpcRequest(req.body);
			if (!validationResult.success) {
				res.status(400).json(validationResult.error);
				logError(new Error('Request validation failed'), {
					event: 'http_streamable_validation_failed',
					sessionId,
					userId: userContext.userId,
				});
				return;
			}

			// 8. Convert Express req/res to Node.js IncomingMessage/ServerResponse
			// Express Request extends IncomingMessage, Response extends ServerResponse
			const nodeReq = req as unknown as IncomingMessage;
			const nodeRes = res as unknown as ServerResponse;

			// 9. Delegate to MCP SDK transport
			// Wrap in request context so server handlers can access session ID
			try {
				await runInContext(
					{ 
						sessionId, 
						userId: userContext.userId, 
						username: userContext.username 
					},
					async () => {
						await transport.handleRequest(nodeReq, nodeRes, validationResult.data);
					}
				);
				
				const duration = Date.now() - startTime;
				logHttpTransport('request_handled', {
					sessionId,
					method: req.method,
					path: req.path,
					duration,
					userId: userContext.userId,
				});
			} catch (error) {
				logError(error as Error, {
					context: 'mcp_request_handling',
					sessionId,
					userId: userContext.userId,
					path: req.path,
				});
				throw error; // Re-throw to be caught by outer catch
			}
		} catch (error) {
			const duration = Date.now() - startTime;
			logError(error as Error, {
				context: 'http_streamable_transport',
				path: req.path,
				method: req.method,
				sessionId,
				duration,
				ip: req.ip,
			});

			// Only send response if not already sent
			if (!res.headersSent) {
				res.status(500).json({
					error: {
						code: -32000,
						message: 'Internal server error',
						data: {
							message: error instanceof Error ? error.message : 'Unknown error',
						},
					},
				});
			}
		}
	}

	/**
	 * Clean up all transports
	 */
	async close(): Promise<void> {
		for (const [sessionId, transport] of this.transports) {
			try {
				await transport.close();
			} catch (error) {
				logError(error as Error, {
					context: 'transport_close',
					sessionId,
				});
			}
		}
		this.transports.clear();
	}
}
