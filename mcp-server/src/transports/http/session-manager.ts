import { v4 as uuidv4 } from 'uuid';
import { config } from '../../config/index.js';
import type { UserContext } from '../../auth/token-validator.js';
import { logHttpTransport, logError } from '../../utils/logger.js';

/**
 * Session state
 */
export type SessionState = 'created' | 'active' | 'orphaned' | 'terminated';

/**
 * Transport type
 */
export type TransportType = 'http-streamable' | 'sse';

/**
 * Client information
 */
export interface ClientInfo {
	userAgent?: string;
	mcpVersion?: string;
	ipAddress?: string;
}

/**
 * Active MCP session
 */
export interface Session {
	id: string;
	token: string;
	userContext: UserContext;
	transport: TransportType;
	state: SessionState;
	createdAt: Date;
	lastActivity: Date;
	clientInfo?: ClientInfo;
}

/**
 * Session manager for HTTP transport
 */
export class SessionManager {
	private sessions = new Map<string, Session>();
	private sessionsByToken = new Map<string, Set<string>>();
	private cleanupInterval: NodeJS.Timeout | null = null;
	private metrics = {
		totalCreated: 0,
		totalTerminated: 0,
	};

	constructor() {
		this.startCleanupInterval();
	}

	/**
	 * Create a new session
	 */
	createSession(
		token: string,
		userContext: UserContext,
		transport: TransportType,
		clientInfo?: ClientInfo
	): Session {
		const sessionId = uuidv4();
		const now = new Date();

		const session: Session = {
			id: sessionId,
			token,
			userContext,
			transport,
			state: 'created',
			createdAt: now,
			lastActivity: now,
			...(clientInfo && { clientInfo }),
		};

		// Store session
		this.sessions.set(sessionId, session);

		// Track by token
		if (!this.sessionsByToken.has(token)) {
			this.sessionsByToken.set(token, new Set());
		}
		this.sessionsByToken.get(token)!.add(sessionId);

		// Update metrics
		this.metrics.totalCreated++;

		logHttpTransport('session_created', {
			sessionId,
			transport,
			userId: userContext.userId,
			username: userContext.username,
		});

		return session;
	}

	/**
	 * Get session by ID
	 */
	getSession(sessionId: string): Session | null {
		const session = this.sessions.get(sessionId);
		if (!session) {
			return null;
		}

		// Check if session is terminated
		if (session.state === 'terminated') {
			return null;
		}

		return session;
	}

	/**
	 * Update session activity timestamp
	 */
	updateActivity(sessionId: string): void {
		const session = this.sessions.get(sessionId);
		if (!session) {
			return;
		}

		session.lastActivity = new Date();

		// Mark as active if it was created
		if (session.state === 'created') {
			session.state = 'active';
		}
	}

	/**
	 * Mark session as orphaned (connection lost)
	 */
	markOrphaned(sessionId: string): void {
		const session = this.sessions.get(sessionId);
		if (!session) {
			return;
		}

		session.state = 'orphaned';
		logHttpTransport('disconnect', {
			sessionId,
			transport: session.transport,
			duration: Date.now() - session.createdAt.getTime(),
		});
	}

	/**
	 * Terminate session and clean up
	 */
	terminateSession(sessionId: string): void {
		const session = this.sessions.get(sessionId);
		if (!session) {
			return;
		}

		session.state = 'terminated';

		// Remove from token tracking
		const tokenSessions = this.sessionsByToken.get(session.token);
		if (tokenSessions) {
			tokenSessions.delete(sessionId);
			if (tokenSessions.size === 0) {
				this.sessionsByToken.delete(session.token);
			}
		}

		// Remove from sessions map
		this.sessions.delete(sessionId);

		// Update metrics
		this.metrics.totalTerminated++;

		logHttpTransport('session_cleanup', {
			sessionId,
			transport: session.transport,
			duration: Date.now() - session.createdAt.getTime(),
		});
	}

	/**
	 * Get all active sessions for a token
	 */
	getSessionsByToken(token: string): Session[] {
		const sessionIds = this.sessionsByToken.get(token);
		if (!sessionIds) {
			return [];
		}

		return Array.from(sessionIds)
			.map((id) => this.sessions.get(id))
			.filter((session): session is Session => session !== undefined && session.state !== 'terminated');
	}

	/**
	 * Get all sessions
	 */
	getAllSessions(): Session[] {
		return Array.from(this.sessions.values()).filter((session) => session.state !== 'terminated');
	}

	/**
	 * Get session statistics
	 */
	getStats(): {
		total: number;
		active: number;
		orphaned: number;
		byTransport: Record<TransportType, number>;
	} {
		const sessions = this.getAllSessions();

		return {
			total: sessions.length,
			active: sessions.filter((s) => s.state === 'active').length,
			orphaned: sessions.filter((s) => s.state === 'orphaned').length,
			byTransport: {
				'http-streamable': sessions.filter((s) => s.transport === 'http-streamable').length,
				'sse': sessions.filter((s) => s.transport === 'sse').length,
			},
		};
	}

	/**
	 * Clean up stale sessions
	 */
	cleanupStaleSessions(): void {
		const now = Date.now();
		const idleTimeoutMs = config.session.idleTimeoutMinutes * 60 * 1000;
		const orphanedTimeoutMs = config.session.orphanedTimeoutSeconds * 1000;

		let cleanedCount = 0;

		for (const session of this.sessions.values()) {
			const idleTime = now - session.lastActivity.getTime();

			// Clean up orphaned sessions after short timeout
			if (session.state === 'orphaned' && idleTime > orphanedTimeoutMs) {
				this.terminateSession(session.id);
				cleanedCount++;
				continue;
			}

			// Clean up idle sessions
			if (session.state === 'active' && idleTime > idleTimeoutMs) {
				this.markOrphaned(session.id);
				this.terminateSession(session.id);
				cleanedCount++;
				continue;
			}
		}

		if (cleanedCount > 0) {
			logHttpTransport('session_cleanup', {
				cleanedSessions: cleanedCount,
				remainingSessions: this.sessions.size,
			});
		}
	}

	/**
	 * Start automatic cleanup interval
	 */
	private startCleanupInterval(): void {
		const intervalMs = config.session.cleanupIntervalSeconds * 1000;

		this.cleanupInterval = setInterval(() => {
			try {
				this.cleanupStaleSessions();
			} catch (error) {
				logError(error as Error, { context: 'session-cleanup-interval' });
			}
		}, intervalMs);

		logHttpTransport('session_created', {
			message: 'Session cleanup interval started',
			intervalSeconds: config.session.cleanupIntervalSeconds,
		});
	}

	/**
	 * Stop cleanup interval
	 */
	stopCleanupInterval(): void {
		if (this.cleanupInterval) {
			clearInterval(this.cleanupInterval);
			this.cleanupInterval = null;
		}
	}

	/**
	 * Shutdown - terminate all sessions and stop cleanup
	 */
	async shutdown(): Promise<void> {
		this.stopCleanupInterval();

		// Terminate all sessions
		const sessionIds = Array.from(this.sessions.keys());
		for (const sessionId of sessionIds) {
			this.terminateSession(sessionId);
		}

		logHttpTransport('session_cleanup', {
			message: 'Session manager shutdown complete',
		});
	}

	/**
	 * Get session metrics
	 */
	getMetrics(): {
		activeSessions: number;
		totalCreated: number;
		totalTerminated: number;
	} {
		return {
			activeSessions: this.sessions.size,
			totalCreated: this.metrics.totalCreated,
			totalTerminated: this.metrics.totalTerminated,
		};
	}
}

/**
 * Singleton session manager instance
 */
let sessionManagerInstance: SessionManager | null = null;

/**
 * Get the singleton session manager instance
 */
export function getSessionManager(): SessionManager {
	if (!sessionManagerInstance) {
		sessionManagerInstance = new SessionManager();
	}
	return sessionManagerInstance;
}
