import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { Authenticator } from '../../../src/auth/authenticator.js';
import { VikunjaClient } from '../../../src/vikunja/client.js';
import { AuthenticationError } from '../../../src/utils/errors.js';
import { VikunjaUser } from '../../../src/vikunja/types.js';

// Mock VikunjaClient
vi.mock('../../../src/vikunja/client.js');

// Mock logger
vi.mock('../../../src/utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

describe('Authenticator', () => {
  let authenticator: Authenticator;
  let mockClient: any;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Create mock VikunjaClient instance
    mockClient = {
      setToken: vi.fn(),
      get: vi.fn(),
    };
    
    // Mock VikunjaClient constructor to return our mock
    vi.mocked(VikunjaClient).mockImplementation(() => mockClient);
    
    authenticator = new Authenticator();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('validateToken', () => {
    it('should validate token and return user context', async () => {
      const mockUser: VikunjaUser = {
        id: 123,
        username: 'testuser',
        email: 'test@example.com',
        name: 'Test User',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValue(mockUser);
      
      const result = await authenticator.validateToken('valid-token');
      
      expect(mockClient.setToken).toHaveBeenCalledWith('valid-token');
      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/user');
      expect(result).toEqual({
        userId: 123,
        username: 'testuser',
        email: 'test@example.com',
        token: 'valid-token',
      });
    });

    it('should cache valid tokens', async () => {
      const mockUser: VikunjaUser = {
        id: 456,
        username: 'cacheduser',
        email: 'cached@example.com',
        name: 'Cached User',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValue(mockUser);
      
      // First call
      const result1 = await authenticator.validateToken('cache-token');
      expect(mockClient.get).toHaveBeenCalledTimes(1);
      
      // Second call - should use cache
      const result2 = await authenticator.validateToken('cache-token');
      expect(mockClient.get).toHaveBeenCalledTimes(1); // Still 1, not 2
      
      // Results should be the same
      expect(result1).toEqual(result2);
    });

    it('should reject invalid tokens', async () => {
      mockClient.get.mockRejectedValue(new Error('Unauthorized'));
      
      await expect(authenticator.validateToken('invalid-token')).rejects.toThrow(
        AuthenticationError
      );
      await expect(authenticator.validateToken('invalid-token')).rejects.toThrow(
        'Invalid or expired token'
      );
    });

    it('should handle Vikunja API errors', async () => {
      const apiError = new Error('Vikunja API is down');
      mockClient.get.mockRejectedValue(apiError);
      
      await expect(authenticator.validateToken('test-token')).rejects.toThrow(
        AuthenticationError
      );
    });

    it('should expire cached tokens after 5 minutes', async () => {
      const mockUser: VikunjaUser = {
        id: 789,
        username: 'expireuser',
        email: 'expire@example.com',
        name: 'Expire User',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValue(mockUser);
      
      // First call
      await authenticator.validateToken('expire-token');
      expect(mockClient.get).toHaveBeenCalledTimes(1);
      
      // Mock time passing (5 minutes + 1 second)
      const originalNow = Date.now;
      Date.now = vi.fn(() => originalNow() + 5 * 60 * 1000 + 1000);
      
      // Second call - cache should be expired
      await authenticator.validateToken('expire-token');
      expect(mockClient.get).toHaveBeenCalledTimes(2);
      
      // Restore Date.now
      Date.now = originalNow;
    });
  });

  describe('invalidateToken', () => {
    it('should remove token from cache', async () => {
      const mockUser: VikunjaUser = {
        id: 111,
        username: 'invalidateuser',
        email: 'invalidate@example.com',
        name: 'Invalidate User',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValue(mockUser);
      
      // Add token to cache
      await authenticator.validateToken('token-to-invalidate');
      expect(mockClient.get).toHaveBeenCalledTimes(1);
      
      // Invalidate token
      authenticator.invalidateToken('token-to-invalidate');
      
      // Next call should hit API again
      await authenticator.validateToken('token-to-invalidate');
      expect(mockClient.get).toHaveBeenCalledTimes(2);
    });
  });

  describe('clearCache', () => {
    it('should clear all cached tokens', async () => {
      const mockUser1: VikunjaUser = {
        id: 1,
        username: 'user1',
        email: 'user1@example.com',
        name: 'User 1',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      const mockUser2: VikunjaUser = {
        id: 2,
        username: 'user2',
        email: 'user2@example.com',
        name: 'User 2',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValueOnce(mockUser1).mockResolvedValueOnce(mockUser2);
      
      // Cache two tokens
      await authenticator.validateToken('token1');
      await authenticator.validateToken('token2');
      expect(mockClient.get).toHaveBeenCalledTimes(2);
      
      // Clear cache
      authenticator.clearCache();
      
      // Reset mock
      mockClient.get.mockResolvedValueOnce(mockUser1).mockResolvedValueOnce(mockUser2);
      
      // Next calls should hit API again
      await authenticator.validateToken('token1');
      await authenticator.validateToken('token2');
      expect(mockClient.get).toHaveBeenCalledTimes(4); // 2 initial + 2 after clear
    });
  });

  describe('Token caching behavior', () => {
    it('should not cache invalid tokens', async () => {
      mockClient.get.mockRejectedValue(new Error('Invalid token'));
      
      // First attempt
      await expect(authenticator.validateToken('bad-token')).rejects.toThrow(
        AuthenticationError
      );
      
      // Second attempt should still call API (not cached)
      await expect(authenticator.validateToken('bad-token')).rejects.toThrow(
        AuthenticationError
      );
      
      expect(mockClient.get).toHaveBeenCalledTimes(2);
    });

    it('should cache different tokens separately', async () => {
      const mockUser1: VikunjaUser = {
        id: 1,
        username: 'user1',
        email: 'user1@example.com',
        name: 'User 1',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      const mockUser2: VikunjaUser = {
        id: 2,
        username: 'user2',
        email: 'user2@example.com',
        name: 'User 2',
        created: '2023-01-01T00:00:00Z',
        updated: '2023-01-01T00:00:00Z',
      };
      
      mockClient.get.mockResolvedValueOnce(mockUser1).mockResolvedValueOnce(mockUser2);
      
      // Validate two different tokens
      const result1 = await authenticator.validateToken('token1');
      const result2 = await authenticator.validateToken('token2');
      
      expect(result1.userId).toBe(1);
      expect(result2.userId).toBe(2);
      expect(mockClient.get).toHaveBeenCalledTimes(2);
      
      // Validate again - should use cache
      await authenticator.validateToken('token1');
      await authenticator.validateToken('token2');
      expect(mockClient.get).toHaveBeenCalledTimes(2); // Still 2
    });
  });
});
