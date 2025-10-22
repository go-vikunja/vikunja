import { describe, it, expect, beforeAll, afterAll, beforeEach, afterEach, vi } from 'vitest';
import type { Server } from 'node:http';
import express, { type Express } from 'express';
import supertest from 'supertest';
import { VikunjaClient } from '../../src/vikunja/client.js';
import { VikunjaMCPServer } from '../../src/server.js';
import { HTTPStreamableTransport } from '../../src/transports/http/http-streamable.js';
import { TokenValidator } from '../../src/auth/token-validator.js';
import { SessionManager } from '../../src/transports/http/session-manager.js';
import { RateLimiter } from '../../src/ratelimit/limiter.js';
import { Authenticator } from '../../src/auth/authenticator.js';
import type { UserContext as ServerUserContext } from '../../src/auth/types.js';
import type { UserContext as TokenUserContext } from '../../src/auth/token-validator.js';

/**
 * REGRESSION FOUND: There are two different UserContext types!
 * - auth/types.ts: has 'token' field, no 'permissions' or 'validatedAt'
 * - auth/token-validator.ts: has 'permissions' and 'validatedAt', no 'token'
 * 
 * This needs to be unified in a future task. For now, we'll create a combined type.
 */
type CombinedUserContext = ServerUserContext & TokenUserContext;

/**
 * Integration Tests for HTTP Transport (T028b)
 * 
 * These tests verify the complete end-to-end flow of the HTTP transport:
 * - Real Express server with all middleware
 * - Actual MCP protocol message exchange
 * - Tool listing and execution
 * - Session management across requests
 * - Error handling and recovery
 * 
 * Unlike unit tests, these tests use the real server stack (with mocked external dependencies).
 */

// Mock ioredis before any imports
vi.mock('ioredis', () => {
	const mockRedis = {
		get: vi.fn().mockResolvedValue(null),
		set: vi.fn().mockResolvedValue('OK'),
		setex: vi.fn().mockResolvedValue('OK'),
		del: vi.fn().mockResolvedValue(1),
		expire: vi.fn().mockResolvedValue(1),
		ttl: vi.fn().mockResolvedValue(300),
		exists: vi.fn().mockResolvedValue(0),
		incr: vi.fn().mockResolvedValue(1),
		ping: vi.fn().mockResolvedValue('PONG'),
		quit: vi.fn().mockResolvedValue('OK'),
		disconnect: vi.fn(),
		on: vi.fn(),
	};

	return {
		default: vi.fn(() => mockRedis),
	};
});

// Mock axios for Vikunja API calls
vi.mock('axios', () => {
	const mockAxios = {
		create: vi.fn(() => mockAxios),
		get: vi.fn(),
		post: vi.fn(),
		put: vi.fn(),
		delete: vi.fn(),
		interceptors: {
			request: {
				use: vi.fn(),
				eject: vi.fn(),
			},
			response: {
				use: vi.fn(),
				eject: vi.fn(),
			},
		},
	};
	return { default: mockAxios };
});

describe('HTTP Transport End-to-End Integration Tests', () => {
	let app: Express;
	let server: Server;
	let transport: HTTPStreamableTransport;
	let tokenValidator: TokenValidator;
	let sessionManager: SessionManager;
	let rateLimiter: RateLimiter;
	let mcpServer: VikunjaMCPServer;

	const mockUserContext: CombinedUserContext = {
		userId: 1,
		username: 'testuser',
		email: 'test@example.com',
		token: 'valid-test-token',
		permissions: ['task:read', 'task:write', 'project:read', 'project:write'],
		validatedAt: new Date(),
	};

	/**
	 * Setup test server with all components
	 */
	beforeAll(async () => {
		// Create dependencies
		const authenticator = new Authenticator();
		const vikunjaClient = new VikunjaClient();

		// Mock Vikunja API responses via generic HTTP methods
		vi.spyOn(vikunjaClient, 'get').mockImplementation(async (path: string) => {
			if (path.includes('/tasks')) {
				return [
					{ id: 1, title: 'Test Task 1', done: false, project_id: 1 },
					{ id: 2, title: 'Test Task 2', done: true, project_id: 1 },
				] as any;
			}
			if (path.includes('/user')) {
				return {
					id: mockUserContext.userId,
					username: mockUserContext.username,
					email: mockUserContext.email,
				} as any;
			}
			return {} as any;
		});

		vi.spyOn(vikunjaClient, 'post').mockImplementation(async (path: string, data: any) => {
			if (path.includes('/tasks')) {
				return {
					id: 3,
					title: data?.title || 'New Task',
					done: false,
					project_id: data?.project_id || 1,
				} as any;
			}
			return {} as any;
		});

		// Create real components
		tokenValidator = new TokenValidator();
		sessionManager = new SessionManager();

		// Create rate limiter with mocked Redis
		const { default: Redis } = await import('ioredis');
		const redis = new Redis();
		rateLimiter = new RateLimiter({
			get: redis.get.bind(redis),
			set: redis.set.bind(redis),
			setex: redis.setex.bind(redis),
			del: redis.del.bind(redis),
			expire: redis.expire.bind(redis),
			ttl: redis.ttl.bind(redis),
			exists: redis.exists.bind(redis),
			incr: redis.incr.bind(redis),
			ping: redis.ping.bind(redis),
		} as any);

		// Create MCP server
		mcpServer = new VikunjaMCPServer(authenticator, rateLimiter, vikunjaClient);

		// Create HTTP transport
		transport = new HTTPStreamableTransport({
			mcpServer,
			sessionManager,
			tokenValidator,
			rateLimiter,
		});

		// Create Express app
		app = express();
		app.use(express.json());

		// Mount MCP endpoint
		app.post('/mcp', (req, res) => {
			void transport.handleRequest(req, res);
		});

		// Start server
		await new Promise<void>((resolve) => {
			server = app.listen(0, () => {
				resolve();
			});
		});
	});

	afterAll(async () => {
		if (server) {
			await new Promise<void>((resolve) => {
				server.close(() => resolve());
			});
		}
		if (transport) {
			await transport.close();
		}
	});

	/**
	 * Helper to create supertest request with required MCP headers
	 * Note: HTTP Streamable protocol always returns SSE format, not plain JSON
	 */
	const mcpRequest = () => supertest(app)
		.post('/mcp')
		.set('Accept', 'application/json, text/event-stream')
		.set('Content-Type', 'application/json');

	/**
	 * Parse SSE (Server-Sent Events) response from MCP HTTP Streamable transport
	 * The SDK always returns responses in SSE format: "event: message\ndata: {...}\n\n"
	 * This is per the MCP HTTP Streamable protocol specification.
	 */
	const parseSSEResponse = (response: supertest.Response) => {
		const text = response.text;
		const lines = text.split('\n');
		const dataLines = lines.filter(line => line.startsWith('data: '));
		if (dataLines.length === 0) {
			throw new Error('No SSE data found in response');
		}
		const jsonStr = dataLines[0].substring(6); // Remove "data: " prefix
		return JSON.parse(jsonStr);
	};

	beforeEach(() => {
		// Mock token validation to return valid user context
		vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);

		// Mock rate limiter to allow requests
		vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	describe('MCP Protocol Connection Flow', () => {
		it('should complete full initialization handshake', async () => {
			const response = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: {
							name: 'test-client',
							version: '1.0.0',
						},
						capabilities: {},
					},
				});

			expect(response.status).toBe(200);
			expect(response.headers['content-type']).toBe('text/event-stream');
			
			const body = parseSSEResponse(response);
			expect(body).toHaveProperty('jsonrpc', '2.0');
			expect(body).toHaveProperty('id', 1);
			expect(body).toHaveProperty('result');

			const result = body.result;
			expect(result).toHaveProperty('protocolVersion', '2024-11-05');
			expect(result).toHaveProperty('serverInfo');
			expect(result.serverInfo).toMatchObject({
				name: 'vikunja-mcp',
				version: '1.0.0',
			});
			expect(result).toHaveProperty('capabilities');
		});

		it('should return session ID in response headers', async () => {
			const response = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			expect(response.status).toBe(200);
			expect(response.headers['x-session-id']).toBeDefined();
			expect(typeof response.headers['x-session-id']).toBe('string');
			expect(response.headers['x-session-id'].length).toBeGreaterThan(0);
		});

		it('should accept initialized notification', async () => {
			// First initialize
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// Then send initialized notification
			const notificationResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					method: 'notifications/initialized',
					params: {},
				});

			// Notifications don't return responses, but should be accepted
			expect(notificationResponse.status).toBe(200);
		});
	});

	describe('Tool Listing', () => {
		it('should return complete list of available tools', async () => {
			// Initialize session first
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// List tools
			const toolsResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 2,
					method: 'tools/list',
					params: {},
				});

			expect(toolsResponse.status).toBe(200);
			expect(toolsResponse.body).toHaveProperty('result');
			expect(toolsResponse.body.result).toHaveProperty('tools');

			const tools = toolsResponse.body.result.tools;
			expect(Array.isArray(tools)).toBe(true);
			expect(tools.length).toBeGreaterThan(0);

			// Verify expected tools are present
			const toolNames = tools.map((t: any) => t.name);
			expect(toolNames).toContain('create_task');
			expect(toolNames).toContain('update_task');
			expect(toolNames).toContain('delete_task');
			expect(toolNames).toContain('create_project');
			expect(toolNames).toContain('update_project');
		});

		it('should include tool schemas in tool list', async () => {
			// Initialize and get session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// List tools
			const toolsResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 2,
					method: 'tools/list',
					params: {},
				});

			const tools = toolsResponse.body.result.tools;
			const createTaskTool = tools.find((t: any) => t.name === 'create_task');

			expect(createTaskTool).toBeDefined();
			expect(createTaskTool).toHaveProperty('name', 'create_task');
			expect(createTaskTool).toHaveProperty('description');
			expect(createTaskTool).toHaveProperty('inputSchema');
			expect(createTaskTool.inputSchema).toHaveProperty('type', 'object');
			expect(createTaskTool.inputSchema).toHaveProperty('properties');
		});
	});

	describe('Tool Execution', () => {
		it('should execute tool with valid arguments', async () => {
			// Initialize and get session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// Set user context in MCP server for this session
			mcpServer.setUserContext('http-session', mockUserContext);

			// Call get_tasks tool
			const toolResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'tools/call',
					params: {
						name: 'get_tasks',
						arguments: {
							project_id: 1,
						},
					},
				});

			expect(toolResponse.status).toBe(200);
			expect(toolResponse.body).toHaveProperty('result');
			expect(toolResponse.body.result).toHaveProperty('content');
			expect(Array.isArray(toolResponse.body.result.content)).toBe(true);
			expect(toolResponse.body.result.content[0]).toHaveProperty('type', 'text');

			// Verify response contains task data
			const resultText = toolResponse.body.result.content[0].text;
			expect(resultText).toContain('Test Task 1');
			expect(resultText).toContain('Test Task 2');
		});

		it('should reject tool call without user context', async () => {
			// Initialize session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// DON'T set user context - simulate unauthorized state

			// Try to call tool
			const toolResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'tools/call',
					params: {
						name: 'get_tasks',
						arguments: { project_id: 1 },
					},
				});

			// Should return error
			expect(toolResponse.status).toBe(200); // JSON-RPC errors are 200
			expect(toolResponse.body).toHaveProperty('error');
			expect(toolResponse.body.error.message).toContain('Unauthorized');
		});

		it('should validate tool arguments against schema', async () => {
			// Initialize session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;
			mcpServer.setUserContext('http-session', mockUserContext);

			// Call tool with invalid arguments (missing required field)
			const toolResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'tools/call',
					params: {
						name: 'create_task',
						arguments: {
							// Missing required 'title' field
							project_id: 1,
						},
					},
				});

			// Should return validation error
			expect(toolResponse.status).toBe(200);
			expect(toolResponse.body).toHaveProperty('error');
			expect(toolResponse.body.error.message).toMatch(/invalid|required|title/i);
		});
	});

	describe('Session Persistence', () => {
		it('should maintain session across multiple requests', async () => {
			// Initialize and get session ID
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;

			// Make second request with same session ID
			const request2 = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 2,
					method: 'tools/list',
					params: {},
				});

			expect(request2.status).toBe(200);
			expect(request2.headers['x-session-id']).toBe(sessionId);

			// Make third request with same session ID
			const request3 = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'ping',
					params: {},
				});

			expect(request3.status).toBe(200);
			expect(request3.headers['x-session-id']).toBe(sessionId);
		});

		it('should track session activity on each request', async () => {
			// Initialize session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;
			const session1 = sessionManager.getSession(sessionId);
			expect(session1).toBeDefined();
			const activity1 = session1!.lastActivity;

			// Wait a bit
			await new Promise((resolve) => setTimeout(resolve, 10));

			// Make another request
			await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 2,
					method: 'tools/list',
					params: {},
				});

			const session2 = sessionManager.getSession(sessionId);
			expect(session2).toBeDefined();
			const activity2 = session2!.lastActivity;

			// Activity should have been updated
			expect(activity2.getTime()).toBeGreaterThan(activity1.getTime());
		});

		it('should create new session if provided session ID is invalid', async () => {
			const response = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', 'invalid-session-id-12345')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			expect(response.status).toBe(200);

			const newSessionId = response.headers['x-session-id'] as string;
			expect(newSessionId).toBeDefined();
			expect(newSessionId).not.toBe('invalid-session-id-12345');

			// Verify new session was created
			const session = sessionManager.getSession(newSessionId);
			expect(session).toBeDefined();
		});
	});

	describe('Error Handling', () => {
		it('should handle Vikunja API errors gracefully', async () => {
			// Mock Vikunja client to throw error
			const vikunjaClient = (mcpServer as any).toolRegistry.vikunjaClient;
			vi.spyOn(vikunjaClient, 'getTasks').mockRejectedValue(
				new Error('Vikunja API connection failed')
			);

			// Initialize session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;
			mcpServer.setUserContext('http-session', mockUserContext);

			// Try to call tool
			const toolResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'tools/call',
					params: {
						name: 'get_tasks',
						arguments: { project_id: 1 },
					},
				});

			// Should return error response
			expect(toolResponse.status).toBe(200);
			expect(toolResponse.body).toHaveProperty('error');
			expect(toolResponse.body.error.message).toContain('Vikunja API');
		});

		it('should handle invalid JSON-RPC requests', async () => {
			const response = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					// Missing jsonrpc field
					id: 1,
					method: 'initialize',
				});

			// Should handle gracefully
			expect([400, 401, 500]).toContain(response.status);
		});

		it('should handle unknown tool names', async () => {
			// Initialize session
			const initResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			const sessionId = initResponse.headers['x-session-id'] as string;
			mcpServer.setUserContext('http-session', mockUserContext);

			// Call non-existent tool
			const toolResponse = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', sessionId)
				.send({
					jsonrpc: '2.0',
					id: 3,
					method: 'tools/call',
					params: {
						name: 'nonexistent_tool',
						arguments: {},
					},
				});

			expect(toolResponse.status).toBe(200);
			expect(toolResponse.body).toHaveProperty('error');
			expect(toolResponse.body.error.message).toMatch(/not found|unknown/i);
		});
	});

	describe('Rate Limiting Integration', () => {
		it('should enforce rate limits across requests', async () => {
			// Mock rate limiter to reject requests
			const rateLimitError = new Error('Rate limit exceeded');
			rateLimitError.name = 'RateLimitError';
			vi.spyOn(rateLimiter, 'checkLimit').mockRejectedValue(rateLimitError);

			const response = await mcpRequest()
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
						capabilities: {},
					},
				});

			expect(response.status).toBe(429);
			const body = parseSSEResponse(response);
			expect(body).toHaveProperty('error');
			expect(body.error.code).toBe(-32003);
			expect(body.error.message).toContain('Rate limit');
			expect(body.error.data).toHaveProperty('retryAfter');
		});
	});
});
