import { RedisStorage } from './storage.js';
import { config } from '../config/index.js';
import { RateLimitError } from '../utils/errors.js';
import { logger } from '../utils/logger.js';

/**
 * Sliding window rate limiter
 */
export class RateLimiter {
  private readonly storage: RedisStorage;
  private readonly windowMs: number = 60 * 1000; // 60 seconds
  private readonly limit: number;
  private readonly burst: number;
  private readonly adminTokens: Set<string> = new Set();

  constructor(storage: RedisStorage) {
    this.storage = storage;
    this.limit = config.rateLimits.default;
    this.burst = config.rateLimits.burst;
  }

  /**
   * Mark token as admin (bypass rate limiting)
   */
  markAsAdmin(token: string): void {
    if (config.rateLimits.adminBypass) {
      this.adminTokens.add(token);
      logger.debug('Token marked as admin', { token: token.substring(0, 8) });
    }
  }

  /**
   * Check if token is admin
   */
  isAdminToken(token: string): boolean {
    return this.adminTokens.has(token);
  }

  /**
   * Check rate limit for token
   */
  async checkLimit(token: string): Promise<void> {
    // Bypass for admin tokens
    if (this.isAdminToken(token)) {
      return;
    }

    const key = `ratelimit:${token}`;
    const now = Date.now();
    const windowStart = now - this.windowMs;

    // Remove old entries outside the window
    await this.storage.zremrangebyscore(key, 0, windowStart);

    // Count requests in current window
    const count = await this.storage.zcard(key);

    // Check if limit exceeded
    if (count >= this.burst) {
      const resetAt = new Date(now + this.windowMs);
      logger.warn('Rate limit exceeded', {
        token: token.substring(0, 8),
        count,
        limit: this.limit,
      });
      throw new RateLimitError('Rate limit exceeded', 0, resetAt, {
        limit: this.limit,
        burst: this.burst,
      });
    }

    // Add current request
    await this.storage.zadd(key, now, `${now}-${Math.random()}`);

    // Set expiry on the key (cleanup)
    const ttl = Math.ceil(this.windowMs / 1000);
    await this.storage.set(key, '1', ttl);

    logger.debug('Rate limit check passed', {
      token: token.substring(0, 8),
      count: count + 1,
      limit: this.limit,
    });
  }

  /**
   * Get remaining requests for token
   */
  async getRemainingRequests(token: string): Promise<number> {
    // Admin tokens have unlimited requests
    if (this.isAdminToken(token)) {
      return this.burst;
    }

    const key = `ratelimit:${token}`;
    const now = Date.now();
    const windowStart = now - this.windowMs;

    // Clean up and count
    await this.storage.zremrangebyscore(key, 0, windowStart);
    const count = await this.storage.zcard(key);

    return Math.max(0, this.burst - count);
  }
}
