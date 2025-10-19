import { z } from 'zod';
import { VikunjaClient } from '../vikunja/client.js';
import { RateLimiter } from '../ratelimit/limiter.js';
import { UserContext } from '../auth/types.js';
import { VikunjaTask } from '../vikunja/types.js';
import { logger } from '../utils/logger.js';

/**
 * Maximum batch size for bulk operations
 */
const MAX_BATCH_SIZE = 100;

/**
 * Input schemas for bulk operation tools
 */
export const BulkUpdateTasksSchema = z.object({
  task_ids: z.array(z.number().int().positive()).min(1).max(MAX_BATCH_SIZE),
  update_data: z.object({
    title: z.string().min(1).max(500).optional(),
    description: z.string().optional(),
    done: z.boolean().optional(),
    due_date: z.string().nullable().optional(),
    priority: z.number().int().min(0).max(5).optional(),
  }),
});

export const BulkCompleteTasksSchema = z.object({
  task_ids: z.array(z.number().int().positive()).min(1).max(MAX_BATCH_SIZE),
});

export const BulkAssignTasksSchema = z.object({
  task_ids: z.array(z.number().int().positive()).min(1).max(MAX_BATCH_SIZE),
  user_id: z.number().int().positive(),
});

export const BulkAddLabelsSchema = z.object({
  task_ids: z.array(z.number().int().positive()).min(1).max(MAX_BATCH_SIZE),
  label_id: z.number().int().positive(),
});

export type BulkUpdateTasksInput = z.infer<typeof BulkUpdateTasksSchema>;
export type BulkCompleteTasksInput = z.infer<typeof BulkCompleteTasksSchema>;
export type BulkAssignTasksInput = z.infer<typeof BulkAssignTasksSchema>;
export type BulkAddLabelsInput = z.infer<typeof BulkAddLabelsSchema>;

/**
 * Bulk operation error
 */
export interface BulkOperationError {
  taskId: number;
  error: string;
}

/**
 * Tool result for bulk operations
 */
export interface BulkToolResult {
  success: boolean;
  message: string;
  successCount: number;
  failureCount: number;
  errors?: BulkOperationError[];
  tasks?: VikunjaTask[];
  error?: string;
}

/**
 * Bulk operation tools for MCP protocol
 */
export class BulkTools {
  constructor(
    private client: VikunjaClient,
    private rateLimiter: RateLimiter
  ) {}

  /**
   * Update multiple tasks at once
   */
  async bulkUpdateTasks(
    input: BulkUpdateTasksInput,
    userContext: UserContext
  ): Promise<BulkToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      const successTasks: VikunjaTask[] = [];
      const errors: BulkOperationError[] = [];

      // Process each task
      for (const taskId of input.task_ids) {
        try {
          const task = await this.client.post<VikunjaTask>(
            `/api/v1/tasks/${taskId}`,
            input.update_data
          );
          successTasks.push(task);
        } catch (error) {
          errors.push({
            taskId,
            error: error instanceof Error ? error.message : String(error),
          });
        }
      }

      const successCount = successTasks.length;
      const failureCount = errors.length;

      logger.info('Bulk update tasks completed', {
        successCount,
        failureCount,
        userId: userContext.userId,
      });

      return {
        success: failureCount === 0,
        message: `Updated ${successCount} of ${input.task_ids.length} tasks`,
        successCount,
        failureCount,
        tasks: successTasks,
        ...(failureCount > 0 && { errors }),
      };
    } catch (error) {
      logger.error('Failed to bulk update tasks', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to bulk update tasks',
        successCount: 0,
        failureCount: input.task_ids.length,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Mark multiple tasks as complete
   */
  async bulkCompleteTasks(
    input: BulkCompleteTasksInput,
    userContext: UserContext
  ): Promise<BulkToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      const successTasks: VikunjaTask[] = [];
      const errors: BulkOperationError[] = [];

      // Process each task
      for (const taskId of input.task_ids) {
        try {
          const task = await this.client.post<VikunjaTask>(`/api/v1/tasks/${taskId}`, {
            done: true,
          });
          successTasks.push(task);
        } catch (error) {
          errors.push({
            taskId,
            error: error instanceof Error ? error.message : String(error),
          });
        }
      }

      const successCount = successTasks.length;
      const failureCount = errors.length;

      logger.info('Bulk complete tasks completed', {
        successCount,
        failureCount,
        userId: userContext.userId,
      });

      return {
        success: failureCount === 0,
        message: `Completed ${successCount} of ${input.task_ids.length} tasks`,
        successCount,
        failureCount,
        tasks: successTasks,
        ...(failureCount > 0 && { errors }),
      };
    } catch (error) {
      logger.error('Failed to bulk complete tasks', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to bulk complete tasks',
        successCount: 0,
        failureCount: input.task_ids.length,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Assign a user to multiple tasks
   */
  async bulkAssignTasks(
    input: BulkAssignTasksInput,
    userContext: UserContext
  ): Promise<BulkToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      const successTasks: VikunjaTask[] = [];
      const errors: BulkOperationError[] = [];

      // Process each task
      for (const taskId of input.task_ids) {
        try {
          const task = await this.client.put<VikunjaTask>(
            `/api/v1/tasks/${taskId}/assignees`,
            { user_id: input.user_id }
          );
          successTasks.push(task);
        } catch (error) {
          errors.push({
            taskId,
            error: error instanceof Error ? error.message : String(error),
          });
        }
      }

      const successCount = successTasks.length;
      const failureCount = errors.length;

      logger.info('Bulk assign tasks completed', {
        successCount,
        failureCount,
        assignedUserId: input.user_id,
        userId: userContext.userId,
      });

      return {
        success: failureCount === 0,
        message: `Assigned user ${input.user_id} to ${successCount} of ${input.task_ids.length} tasks`,
        successCount,
        failureCount,
        tasks: successTasks,
        ...(failureCount > 0 && { errors }),
      };
    } catch (error) {
      logger.error('Failed to bulk assign tasks', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to bulk assign tasks',
        successCount: 0,
        failureCount: input.task_ids.length,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Add a label to multiple tasks
   */
  async bulkAddLabels(
    input: BulkAddLabelsInput,
    userContext: UserContext
  ): Promise<BulkToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      const successTasks: VikunjaTask[] = [];
      const errors: BulkOperationError[] = [];

      // Process each task
      for (const taskId of input.task_ids) {
        try {
          const task = await this.client.put<VikunjaTask>(`/api/v1/tasks/${taskId}/labels`, {
            label_id: input.label_id,
          });
          successTasks.push(task);
        } catch (error) {
          errors.push({
            taskId,
            error: error instanceof Error ? error.message : String(error),
          });
        }
      }

      const successCount = successTasks.length;
      const failureCount = errors.length;

      logger.info('Bulk add labels completed', {
        successCount,
        failureCount,
        labelId: input.label_id,
        userId: userContext.userId,
      });

      return {
        success: failureCount === 0,
        message: `Added label ${input.label_id} to ${successCount} of ${input.task_ids.length} tasks`,
        successCount,
        failureCount,
        tasks: successTasks,
        ...(failureCount > 0 && { errors }),
      };
    } catch (error) {
      logger.error('Failed to bulk add labels', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to bulk add labels',
        successCount: 0,
        failureCount: input.task_ids.length,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
