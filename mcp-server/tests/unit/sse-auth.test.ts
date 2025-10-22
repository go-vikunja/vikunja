import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Request, Response, NextFunction } from 'express';
import { Authenticator } from '../../src/auth/authenticator.js';
import type { UserContext } from '../../src/auth/types.js';

// Mock implementation of SSE auth middleware (to be implemented in src/transport/http.ts)
function createSSEAuthMiddleware(authenticator: Authenticator) {
  return async (req: Request, res: Response, next: NextFunction) => {
    try {
      // Extract token from Authorization header or query parameter
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

      // Validate token with existing Authenticator (includes 5-min cache)
      const userContext = await authenticator.validateToken(token);

      // Store user context in request for SSE handler
      (req as any).userContext = userContext;

      next();
    } catch (error) {
      res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid or expired authentication token',
      });
    }
  };
}

describe('SSE Auth Middleware', () => {
  let mockAuthenticator: Authenticator;
  let mockRequest: Partial<Request>;
  let mockResponse: Partial<Response>;
  let mockNext: NextFunction;
  let statusMock: any;
  let jsonMock: any;

  beforeEach(() => {
    // Reset mocks
    mockAuthenticator = {
      validateToken: vi.fn(),
      invalidateToken: vi.fn(),
      clearCache: vi.fn(),
    } as any;

    statusMock = vi.fn().mockReturnThis();
    jsonMock = vi.fn();

    mockRequest = {
      headers: {},
      query: {},
      ip: '127.0.0.1',
    };

    mockResponse = {
      status: statusMock,
      json: jsonMock,
    } as any;

    mockNext = vi.fn() as any;
  });

  it('should extract token from Authorization header', async () => {
    // Arrange
    const token = 'valid-token-123';
    const userContext: UserContext = {
      userId: 1,
      username: 'testuser',
      email: 'test@example.com',
      token,
    };

    mockRequest.headers = {
      authorization: `Bearer ${token}`,
    };

    vi.mocked(mockAuthenticator.validateToken).mockResolvedValue(userContext);

    const middleware = createSSEAuthMiddleware(mockAuthenticator);

    // Act
    await middleware(
      mockRequest as Request,
      mockResponse as Response,
      mockNext
    );

    // Assert
    expect(mockAuthenticator.validateToken).toHaveBeenCalledWith(token);
    expect((mockRequest as any).userContext).toEqual(userContext);
    expect(mockNext).toHaveBeenCalled();
    expect(statusMock).not.toHaveBeenCalled();
  });

  it('should extract token from query parameter', async () => {
    // Arrange
    const token = 'query-token-456';
    const userContext: UserContext = {
      userId: 2,
      username: 'queryuser',
      email: 'query@example.com',
      token,
    };

    mockRequest.query = { token };

    vi.mocked(mockAuthenticator.validateToken).mockResolvedValue(userContext);

    const middleware = createSSEAuthMiddleware(mockAuthenticator);

    // Act
    await middleware(
      mockRequest as Request,
      mockResponse as Response,
      mockNext
    );

    // Assert
    expect(mockAuthenticator.validateToken).toHaveBeenCalledWith(token);
    expect((mockRequest as any).userContext).toEqual(userContext);
    expect(mockNext).toHaveBeenCalled();
    expect(statusMock).not.toHaveBeenCalled();
  });

  it('should return 401 for missing token', async () => {
    // Arrange
    mockRequest.headers = {};
    mockRequest.query = {};

    const middleware = createSSEAuthMiddleware(mockAuthenticator);

    // Act
    await middleware(
      mockRequest as Request,
      mockResponse as Response,
      mockNext
    );

    // Assert
    expect(mockAuthenticator.validateToken).not.toHaveBeenCalled();
    expect(statusMock).toHaveBeenCalledWith(401);
    expect(jsonMock).toHaveBeenCalledWith({
      error: 'Unauthorized',
      message: expect.stringContaining('Missing authentication token'),
    });
    expect(mockNext).not.toHaveBeenCalled();
  });

  it('should return 401 for invalid token', async () => {
    // Arrange
    const invalidToken = 'invalid-token';
    mockRequest.headers = {
      authorization: `Bearer ${invalidToken}`,
    };

    vi.mocked(mockAuthenticator.validateToken).mockRejectedValue(
      new Error('Invalid token')
    );

    const middleware = createSSEAuthMiddleware(mockAuthenticator);

    // Act
    await middleware(
      mockRequest as Request,
      mockResponse as Response,
      mockNext
    );

    // Assert
    expect(mockAuthenticator.validateToken).toHaveBeenCalledWith(invalidToken);
    expect(statusMock).toHaveBeenCalledWith(401);
    expect(jsonMock).toHaveBeenCalledWith({
      error: 'Unauthorized',
      message: 'Invalid or expired authentication token',
    });
    expect(mockNext).not.toHaveBeenCalled();
  });

  it('should populate req.userContext on valid token', async () => {
    // Arrange
    const token = 'valid-context-token';
    const userContext: UserContext = {
      userId: 3,
      username: 'contextuser',
      email: 'context@example.com',
      token,
    };

    mockRequest.headers = {
      authorization: `Bearer ${token}`,
    };

    vi.mocked(mockAuthenticator.validateToken).mockResolvedValue(userContext);

    const middleware = createSSEAuthMiddleware(mockAuthenticator);

    // Act
    await middleware(
      mockRequest as Request,
      mockResponse as Response,
      mockNext
    );

    // Assert
    expect((mockRequest as any).userContext).toBeDefined();
    expect((mockRequest as any).userContext.userId).toBe(3);
    expect((mockRequest as any).userContext.username).toBe('contextuser');
    expect((mockRequest as any).userContext.email).toBe('context@example.com');
    expect((mockRequest as any).userContext.token).toBe(token);
    expect(mockNext).toHaveBeenCalled();
  });
});
