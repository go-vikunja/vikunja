import { VikunjaClient } from '../vikunja/client.js';
import { VikunjaUser } from '../vikunja/types.js';
import { UserContext } from './types.js';
import { AuthenticationError } from '../utils/errors.js';
import { logger } from '../utils/logger.js';

/**
 * Token cache entry
 */
interface CacheEntry {
  userContext: UserContext;
  expiresAt: number;
}

/**
 * Authenticator for validating Vikunja API tokens
 */
export class Authenticator {
  private readonly tokenCache: Map<string, CacheEntry> = new Map();
  private readonly cacheExpiryMs: number = 5 * 60 * 1000; // 5 minutes

  /**
   * Validate token and return user context
   */
  async validateToken(token: string): Promise<UserContext> {
    // Check cache first
    const cached = this.tokenCache.get(token);
    if (cached && cached.expiresAt > Date.now()) {
      logger.debug('Token found in cache');
      return cached.userContext;
    }

    // Call Vikunja API to validate token
    try {
      const client = new VikunjaClient();
      client.setToken(token);

      const user = await client.get<VikunjaUser>('/api/v1/user');

      const userContext: UserContext = {
        userId: user.id,
        username: user.username,
        email: user.email,
        token,
      };

      // Cache the result
      this.tokenCache.set(token, {
        userContext,
        expiresAt: Date.now() + this.cacheExpiryMs,
      });

      logger.info(`Token validated for user: ${user.username}`, { userId: user.id });

      return userContext;
    } catch (error) {
      logger.error('Token validation failed', { error });
      throw new AuthenticationError('Invalid or expired token');
    }
  }

  /**
   * Invalidate cached token
   */
  invalidateToken(token: string): void {
    this.tokenCache.delete(token);
    logger.debug('Token invalidated from cache');
  }

  /**
   * Clear all cached tokens
   */
  clearCache(): void {
    this.tokenCache.clear();
    logger.debug('Token cache cleared');
  }
}
