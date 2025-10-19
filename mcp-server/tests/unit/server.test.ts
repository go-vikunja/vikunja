import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VikunjaMCPServer } from '../../src/server.js';
import type { Authenticator } from '../../src/auth/authenticator.js';
import type { RateLimiter } from '../../src/ratelimit/limiter.js';
import type { VikunjaClient } from '../../src/vikunja/client.js';

// Mock dependencies
vi.mock('../../src/auth/authenticator.js');
vi.mock('../../src/ratelimit/limiter.js');
vi.mock('../../src/vikunja/client.js');

describe('VikunjaMCPServer', () => {
  let server: VikunjaMCPServer;
  let mockAuthenticator: Authenticator;
  let mockRateLimiter: RateLimiter;
  let mockVikunjaClient: VikunjaClient;

  beforeEach(() => {
    // Create mock instances
    mockAuthenticator = {
      validateToken: vi.fn(),
      invalidateToken: vi.fn(),
    } as any;

    mockRateLimiter = {
      checkLimit: vi.fn(),
      getRemainingRequests: vi.fn(),
    } as any;

    mockVikunjaClient = {
      setToken: vi.fn(),
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    } as any;

    server = new VikunjaMCPServer(
      mockAuthenticator,
      mockRateLimiter,
      mockVikunjaClient
    );
  });

  describe('initialization', () => {
    it('should respond to initialize request', async () => {
      const result = await server.handleInitialize({
        protocolVersion: '2024-11-05',
        capabilities: {},
        clientInfo: {
          name: 'test-client',
          version: '1.0.0',
        },
      });

      expect(result).toHaveProperty('protocolVersion', '2024-11-05');
      expect(result).toHaveProperty('serverInfo');
      expect(result.serverInfo).toHaveProperty('name', 'vikunja-mcp');
      expect(result.serverInfo).toHaveProperty('version');
      expect(result).toHaveProperty('capabilities');
    });

    it('should declare capabilities', async () => {
      const result = await server.handleInitialize({
        protocolVersion: '2024-11-05',
        capabilities: {},
        clientInfo: {
          name: 'test-client',
          version: '1.0.0',
        },
      });

      expect(result.capabilities).toHaveProperty('resources');
      expect(result.capabilities).toHaveProperty('tools');
      expect(result.capabilities).toHaveProperty('prompts');
    });

    it('should handle initialized notification', async () => {
      // This shouldn't throw
      await expect(server.handleInitialized()).resolves.toBeUndefined();
    });
  });

  describe('authentication', () => {
    it('should authenticate connection with valid token', async () => {
      const mockUser = {
        userId: 1,
        username: 'testuser',
        email: 'test@example.com',
        token: 'valid-token-123',
      };

      vi.mocked(mockAuthenticator.validateToken).mockResolvedValue(mockUser);

      const result = await server.authenticateConnection('valid-token-123');

      expect(result).toEqual(mockUser);
      expect(mockAuthenticator.validateToken).toHaveBeenCalledWith('valid-token-123');
    });

    it('should reject connection with invalid token', async () => {
      vi.mocked(mockAuthenticator.validateToken).mockRejectedValue(
        new Error('Invalid token')
      );

      await expect(
        server.authenticateConnection('invalid-token')
      ).rejects.toThrow('Invalid token');
    });
  });

  describe('user context storage', () => {
    it('should store user context for connection', () => {
      const connectionId = 'conn-123';
      const userContext = {
        userId: 1,
        username: 'testuser',
        email: 'test@example.com',
        token: 'token-123',
      };

      server.setUserContext(connectionId, userContext);
      const retrieved = server.getUserContext(connectionId);

      expect(retrieved).toEqual(userContext);
    });

    it('should return undefined for non-existent connection', () => {
      const retrieved = server.getUserContext('non-existent');
      expect(retrieved).toBeUndefined();
    });

    it('should remove user context', () => {
      const connectionId = 'conn-123';
      const userContext = {
        userId: 1,
        username: 'testuser',
        email: 'test@example.com',
        token: 'token-123',
      };

      server.setUserContext(connectionId, userContext);
      server.removeUserContext(connectionId);

      const retrieved = server.getUserContext(connectionId);
      expect(retrieved).toBeUndefined();
    });
  });
});
