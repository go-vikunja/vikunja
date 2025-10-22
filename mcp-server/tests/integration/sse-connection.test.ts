import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import supertest from 'supertest';
import express, { type Express } from 'express';
import type { UserContext } from '../../src/auth/types.js';
import { Authenticator } from '../../src/auth/authenticator.js';

// Mock SSE endpoint setup (to be implemented in src/transport/http.ts)
function createMockSSEApp() {
  const app = express();
  app.use(express.json());

  const mockAuthenticator = {
    validateToken: vi.fn(),
  } as any;

  // Stub auth middleware
  const authMiddleware = async (req: any, res: any, next: any) => {
    try {
      const authHeader = req.headers.authorization;
      const token =
        authHeader?.replace('Bearer ', '') || (req.query.token as string);

      if (!token) {
        res.status(401).json({
          error: 'Unauthorized',
          message:
            'Missing authentication token. Provide token via Authorization header or ?token= query parameter.',
        });
        return;
      }

      const userContext = await mockAuthenticator.validateToken(token);
      req.userContext = userContext;
      next();
    } catch (error) {
      res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid or expired authentication token',
      });
    }
  };

  // SSE endpoint
  app.post('/sse', authMiddleware, (req: any, res: any) => {
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    // Send connected event
    const connectionId = 'test-connection-id';
    res.write(
      `event: connected\ndata: ${JSON.stringify({ connectionId })}\n\n`
    );

    // Keep connection open briefly for testing
    setTimeout(() => {
      res.end();
    }, 100);
  });

  return { app, mockAuthenticator };
}

describe('SSE Connection Integration', () => {
  let app: Express;
  let mockAuthenticator: any;

  beforeAll(() => {
    const setup = createMockSSEApp();
    app = setup.app;
    mockAuthenticator = setup.mockAuthenticator;
  });

  afterAll(() => {
    vi.clearAllMocks();
  });

  it('should establish SSE connection with valid token in header', async () => {
    // Arrange
    const token = 'valid-token-123';
    const userContext: UserContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token,
    };

    mockAuthenticator.validateToken.mockResolvedValue(userContext);

    // Act
    const response = await supertest(app)
      .post('/sse')
      .set('Authorization', `Bearer ${token}`)
      .expect(200);

    // Assert
    expect(response.headers['content-type']).toContain('text/event-stream');
    expect(response.text).toContain('event: connected');
    expect(response.text).toContain('connectionId');
  });

  it('should return 401 with invalid token', async () => {
    // Arrange
    const invalidToken = 'invalid-token';
    mockAuthenticator.validateToken.mockRejectedValue(
      new Error('Invalid token')
    );

    // Act
    const response = await supertest(app)
      .post('/sse')
      .set('Authorization', `Bearer ${invalidToken}`)
      .expect(401);

    // Assert
    expect(response.body.error).toBe('Unauthorized');
    expect(response.body.message).toContain('Invalid or expired');
  });

  it('should return 401 without token', async () => {
    // Act
    const response = await supertest(app).post('/sse').expect(401);

    // Assert
    expect(response.body.error).toBe('Unauthorized');
    expect(response.body.message).toContain('Missing authentication token');
  });

  it('should receive connected event with connectionId', async () => {
    // Arrange
    const token = 'connected-event-token';
    const userContext: UserContext = {
      userId: 2,
      username: 'eventuser',
      email: 'event@example.com',
      token,
    };

    mockAuthenticator.validateToken.mockResolvedValue(userContext);

    // Act
    const response = await supertest(app)
      .post('/sse')
      .set('Authorization', `Bearer ${token}`)
      .expect(200);

    // Assert
    expect(response.text).toMatch(/event: connected/);
    expect(response.text).toMatch(/"connectionId":\s*"[^"]+"/);
  });

  it('should handle MCP tool execution over SSE', async () => {
    // Note: This is a stub test - actual MCP tool execution will be implemented
    // in the full HTTP transport implementation
    
    // Arrange
    const token = 'tool-execution-token';
    const userContext: UserContext = {
      userId: 3,
      username: 'tooluser',
      email: 'tool@example.com',
      token,
    };

    mockAuthenticator.validateToken.mockResolvedValue(userContext);

    // Act
    const response = await supertest(app)
      .post('/sse')
      .set('Authorization', `Bearer ${token}`)
      .expect(200);

    // Assert - for now, just verify connection is established
    expect(response.headers['content-type']).toContain('text/event-stream');
    expect(response.text).toContain('event: connected');
  });
});

