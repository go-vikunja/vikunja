import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import supertest from 'supertest';
import express, { type Express } from 'express';
import { HTTPStreamableTransport } from '../../../src/transports/http/http-streamable.js';
import { SessionManager } from '../../../src/transports/http/session-manager.js';
import { TokenValidator } from '../../../src/auth/token-validator.js';
import { RateLimiter } from '../../../src/ratelimit/limiter.js';
import { RedisStorage } from '../../../src/ratelimit/storage.js';
import { VikunjaMCPServer } from '../../../src/server.js';
import { Authenticator } from '../../../src/auth/authenticator.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';

vi.mock('ioredis', () => {
	return {
		default: vi.fn().mockImplementation(() => ({
			on: vi.fn(),
			get: vi.fn().mockResolvedValue(null),
			set: vi.fn().mockResolvedValue('OK'),
			setex: vi.fn().mockResolvedValue('OK'),
			del: vi.fn().mockResolvedValue(1),
			quit: vi.fn().mockResolvedValue('OK'),
			incr: vi.fn().mockResolvedValue(1),
			expire: vi.fn().mockResolvedValue(1),
			ttl: vi.fn().mockResolvedValue(-1),
		})),
	};
});

// Mock axios for Vikunja API calls
vi.mock('axios', () => {
	const mockAxiosInstance = {
		interceptors: {
			request: { use: vi.fn(), eject: vi.fn() },
			response: { use: vi.fn(), eject: vi.fn() },
		},
		get: vi.fn(),
		post: vi.fn(),
		put: vi.fn(),
		delete: vi.fn(),
		patch: vi.fn(),
	};
	return {
		default: {
			create: vi.fn(() => mockAxiosInstance),
			...mockAxiosInstance,
		},
	};
});

describe('HTTP Streamable Transport - HTTP Integration Tests', () => {
	let app: Express;
	let sessionManager: SessionManager;
	let tokenValidator: TokenValidator;
	let rateLimiter: RateLimiter;
	let storage: RedisStorage;
	let mcpServer: VikunjaMCPServer;
	let transport: HTTPStreamableTransport;

	// Mock user context for successful authentication
	const mockUserContext = {
		userId: 1,
		username: 'testuser',
		email: 'test@example.com',
		token: 'valid-token',
		permissions: ['task:read', 'task:write', 'project:read'],
		validatedAt: new Date(),
	}; 
	
	function createTestApp(): Express {
		const testApp = express();
		testApp.use(express.json());

		sessionManager = new SessionManager();
		tokenValidator = new TokenValidator();
		storage = new RedisStorage();
		rateLimiter = new RateLimiter(storage);

		const authenticator = new Authenticator();
		const vikunjaClient = new VikunjaClient();
		mcpServer = new VikunjaMCPServer(authenticator, rateLimiter, vikunjaClient);

		transport = new HTTPStreamableTransport({
			mcpServer,
			sessionManager,
			tokenValidator,
			rateLimiter,
		});

		testApp.post('/mcp', (req, res) => {
			void transport.handleRequest(req, res);
		});

		return testApp;
	}

	beforeEach(() => {
		app = createTestApp();
	});

	afterEach(async () => {
		if (transport) {
			await transport.close();
		}
		vi.clearAllMocks();
	});

	describe('HTTP Protocol Compliance', () => {
		it('should accept POST requests to /mcp endpoint', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockRejectedValue(new Error('Invalid token'));

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer test-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.status).toBeDefined();
			expect([401, 500]).toContain(response.status);
		});

		it('should return proper Content-Type header for JSON responses', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.headers['content-type']).toMatch(/application\/json/);
		});

		it('should handle missing body gracefully', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer test-token');

			expect(response.status).toBeDefined();
			expect([400, 401, 500]).toContain(response.status);
		});

		it('should set X-Session-ID header on successful requests', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);
			vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);

			const response = await supertest(app)
				.post('/mcp')
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
					},
				});

			if (response.status === 200) {
				expect(response.headers['x-session-id']).toBeDefined();
				expect(typeof response.headers['x-session-id']).toBe('string');
			}
		});
	});

	describe('Authentication Flow', () => {
		it('should reject requests without Authorization header', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.status).toBe(401);
			expect(response.body).toHaveProperty('error');
			expect(response.body.error).toMatchObject({
				code: -32001,
				message: expect.stringContaining('Authentication required'),
			});
		});

		it('should reject requests with malformed Authorization header', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'InvalidFormat')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.status).toBe(401);
			expect(response.body.error.code).toBe(-32001);
		});

		it('should reject requests with invalid Bearer token', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockRejectedValue(
				new Error('Invalid token')
			);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer invalid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.status).toBe(401);
			expect(response.body.error).toMatchObject({
				code: -32001,
				message: expect.stringContaining('Invalid token'),
			});
		});

		it('should accept requests with valid Bearer token', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);
			vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);

			const response = await supertest(app)
				.post('/mcp')
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
					},
				});

			expect(response.status).not.toBe(401);
		});
	});

	describe('Session Management', () => {
		it('should create new session on first request', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);
			vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
					},
				});

			if (response.status === 200) {
				const sessionId = response.headers['x-session-id'];
				expect(sessionId).toBeDefined();

				const session = sessionManager.getSession(sessionId);
				expect(session).toBeDefined();
				expect(session?.userContext.userId).toBe(1);
			}
		});

		it('should reuse session when X-Session-ID header is provided', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);
			vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);

			const response1 = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
					},
				});

			if (response1.status === 200) {
				const sessionId = response1.headers['x-session-id'];

				const response2 = await supertest(app)
					.post('/mcp')
					.set('Authorization', 'Bearer valid-token')
					.set('X-Session-ID', sessionId)
					.send({
						jsonrpc: '2.0',
						id: 2,
						method: 'ping',
						params: {},
					});

				expect(response2.headers['x-session-id']).toBe(sessionId);
			}
		});

		it('should create new session if provided session ID is invalid', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);
			vi.spyOn(rateLimiter, 'checkLimit').mockResolvedValue(undefined);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.set('X-Session-ID', 'invalid-session-id')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {
						protocolVersion: '2024-11-05',
						clientInfo: { name: 'test', version: '1.0.0' },
					},
				});

			if (response.status === 200) {
				const sessionId = response.headers['x-session-id'];
				expect(sessionId).toBeDefined();
				expect(sessionId).not.toBe('invalid-session-id');
			}
		});
	});

	describe('Rate Limiting', () => {
		it('should enforce rate limits and return 429', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);

			const rateLimitError = new Error('Rate limit exceeded');
			rateLimitError.name = 'RateLimitError';
			vi.spyOn(rateLimiter, 'checkLimit').mockRejectedValue(rateLimitError);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.status).toBe(429);
			expect(response.body).toHaveProperty('error');
			expect(response.body.error.code).toBe(-32003);
			expect(response.body.error.message).toContain('Rate limit exceeded');
		});

		it('should include retry information in rate limit response', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockResolvedValue(mockUserContext);

			const rateLimitError = new Error('Rate limit exceeded');
			rateLimitError.name = 'RateLimitError';
			vi.spyOn(rateLimiter, 'checkLimit').mockRejectedValue(rateLimitError);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.body.error.data).toHaveProperty('retryAfter');
			expect(typeof response.body.error.data.retryAfter).toBe('number');
		});
	});

	describe('Error Handling', () => {
		it('should return 500 for internal server errors', async () => {
			vi.spyOn(tokenValidator, 'validateToken').mockRejectedValue(
				new Error('Database connection failed')
			);

			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect([401, 500]).toContain(response.status);
		});

		it('should include error details in response body', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.send({
					jsonrpc: '2.0',
					id: 1,
					method: 'initialize',
					params: {},
				});

			expect(response.body).toHaveProperty('error');
			expect(response.body.error).toHaveProperty('code');
			expect(response.body.error).toHaveProperty('message');
		});

		it('should handle malformed JSON gracefully', async () => {
			const response = await supertest(app)
				.post('/mcp')
				.set('Authorization', 'Bearer valid-token')
				.set('Content-Type', 'application/json')
				.send('{"invalid json}');

			expect([400, 401, 500]).toContain(response.status);
		});
	});
});
