import type { Response } from 'express';
import type { UserContext } from '../auth/types.js';

/**
 * Represents an active SSE connection
 */
export interface SSEConnection {
  /**
   * Unique connection identifier
   */
  id: string;

  /**
   * Authenticated user context
   */
  userContext: UserContext;

  /**
   * Express response object (for SSE streaming)
   */
  response: Response;

  /**
   * Connection establishment timestamp
   */
  connectedAt: Date;

  /**
   * Last activity timestamp (for idle detection)
   */
  lastActivityAt: Date;

  /**
   * Connection state
   */
  state: 'connected' | 'closing' | 'closed';
}

/**
 * SSE connection manager
 * Tracks all active connections for graceful shutdown
 */
export class SSEConnectionManager {
  private connections: Map<string, SSEConnection>;

  constructor() {
    this.connections = new Map();
  }

  /**
   * Add new connection
   */
  add(connection: SSEConnection): void {
    this.connections.set(connection.id, connection);
  }

  /**
   * Remove connection
   */
  remove(connectionId: string): void {
    this.connections.delete(connectionId);
  }

  /**
   * Get connection by ID
   */
  get(connectionId: string): SSEConnection | undefined {
    return this.connections.get(connectionId);
  }

  /**
   * Get all active connections
   */
  getAll(): SSEConnection[] {
    return Array.from(this.connections.values());
  }

  /**
   * Get connection count
   */
  count(): number {
    return this.connections.size;
  }

  /**
   * Gracefully close all connections
   */
  async closeAll(): Promise<void> {
    const closePromises = Array.from(this.connections.values()).map((conn) => {
      conn.state = 'closing';
      // Send close event via SSE
      conn.response.write('event: close\ndata: {"reason": "server shutdown"}\n\n');
      conn.response.end();
      conn.state = 'closed';
      return Promise.resolve();
    });

    await Promise.all(closePromises);
    this.connections.clear();
  }
}
