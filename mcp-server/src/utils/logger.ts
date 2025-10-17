import winston from 'winston';
import { config } from '../config/index.js';

/**
 * Create Winston logger instance
 */
function createLogger(): winston.Logger {
  const formats = [];

  // Add timestamp
  formats.push(winston.format.timestamp());

  // Add request ID if available
  formats.push(
    winston.format((info) => {
      if (info['requestId']) {
        info['requestId'] = info['requestId'];
      }
      return info;
    })()
  );

  // Format based on config
  if (config.logging.format === 'json') {
    formats.push(winston.format.json());
  } else {
    formats.push(
      winston.format.printf(({ level, message, timestamp, requestId }) => {
        const reqId = requestId ? `[${requestId}] ` : '';
        return `${timestamp} ${level}: ${reqId}${message}`;
      })
    );
  }

  const transports: winston.transport[] = [];

  // Console transport (always enabled in development)
  if (process.env['NODE_ENV'] !== 'production') {
    transports.push(
      new winston.transports.Console({
        format: winston.format.combine(winston.format.colorize(), ...formats),
      })
    );
  }

  // File transport (production)
  if (process.env['NODE_ENV'] === 'production') {
    transports.push(
      new winston.transports.File({
        filename: '/var/log/vikunja-mcp/app.log',
        format: winston.format.combine(...formats),
      })
    );
  }

  // If no transports, add console as fallback
  if (transports.length === 0) {
    transports.push(
      new winston.transports.Console({
        format: winston.format.combine(...formats),
      })
    );
  }

  return winston.createLogger({
    level: config.logging.level,
    transports,
  });
}

/**
 * Singleton logger instance
 */
export const logger = createLogger();

/**
 * Log an incoming request
 */
export function logRequest(requestId: string, method: string, path: string): void {
  logger.info(`${method} ${path}`, { requestId });
}

/**
 * Log an error with context
 */
export function logError(error: Error, context?: Record<string, unknown>): void {
  logger.error(error.message, {
    error: {
      name: error.name,
      message: error.message,
      stack: error.stack,
    },
    ...context,
  });
}

/**
 * Log a tool call
 */
export function logToolCall(
  requestId: string,
  toolName: string,
  params: unknown,
  userId?: string
): void {
  logger.info(`Tool call: ${toolName}`, {
    requestId,
    toolName,
    params,
    userId,
  });
}
