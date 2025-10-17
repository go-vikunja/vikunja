import { z } from 'zod';
import { VikunjaClient } from '../vikunja/client.js';
import { RateLimiter } from '../ratelimit/limiter.js';
import { UserContext } from '../auth/types.js';
import { VikunjaTask, VikunjaLabel } from '../vikunja/types.js';
import { logger } from '../utils/logger.js';

/**
 * Input schemas for assignment tools
 */
export const AssignTaskSchema = z.object({
  task_id: z.number().int().positive(),
  user_id: z.number().int().positive(),
});

export const UnassignTaskSchema = z.object({
  task_id: z.number().int().positive(),
  user_id: z.number().int().positive(),
});

export const AddLabelSchema = z.object({
  task_id: z.number().int().positive(),
  label_id: z.number().int().positive(),
});

export const RemoveLabelSchema = z.object({
  task_id: z.number().int().positive(),
  label_id: z.number().int().positive(),
});

export const CreateLabelSchema = z.object({
  title: z.string().min(1).max(250),
  description: z.string().optional(),
  hex_color: z.string().regex(/^#[0-9a-fA-F]{6}$/).optional(),
});

export type AssignTaskInput = z.infer<typeof AssignTaskSchema>;
export type UnassignTaskInput = z.infer<typeof UnassignTaskSchema>;
export type AddLabelInput = z.infer<typeof AddLabelSchema>;
export type RemoveLabelInput = z.infer<typeof RemoveLabelSchema>;
export type CreateLabelInput = z.infer<typeof CreateLabelSchema>;

/**
 * Tool result for assignment operations
 */
export interface AssignmentToolResult {
  success: boolean;
  message: string;
  task?: VikunjaTask;
  label?: VikunjaLabel;
  error?: string;
}

/**
 * Assignment tools for MCP protocol
 */
export class AssignmentTools {
  constructor(
    private client: VikunjaClient,
    private rateLimiter: RateLimiter
  ) {}

  /**
   * Assign a user to a task
   */
  async assignTask(
    input: AssignTaskInput,
    userContext: UserContext
  ): Promise<AssignmentToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Assign user to task
      const task = await this.client.put<VikunjaTask>(
        `/api/v1/tasks/${input.task_id}/assignees`,
        { user_id: input.user_id }
      );

      logger.info('User assigned to task', {
        taskId: input.task_id,
        assignedUserId: input.user_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `User ${input.user_id} assigned to task "${task.title}"`,
        task,
      };
    } catch (error) {
      logger.error('Failed to assign user to task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to assign user to task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Remove a user from a task
   */
  async unassignTask(
    input: UnassignTaskInput,
    userContext: UserContext
  ): Promise<AssignmentToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Unassign user from task
      const task = await this.client.delete<VikunjaTask>(
        `/api/v1/tasks/${input.task_id}/assignees/${input.user_id}`
      );

      logger.info('User unassigned from task', {
        taskId: input.task_id,
        unassignedUserId: input.user_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `User ${input.user_id} removed from task "${task.title}"`,
        task,
      };
    } catch (error) {
      logger.error('Failed to unassign user from task', {
        error,
        userId: userContext.userId,
      });
      return {
        success: false,
        message: 'Failed to unassign user from task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Add a label to a task
   */
  async addLabel(
    input: AddLabelInput,
    userContext: UserContext
  ): Promise<AssignmentToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Add label to task
      const task = await this.client.put<VikunjaTask>(
        `/api/v1/tasks/${input.task_id}/labels`,
        { label_id: input.label_id }
      );

      logger.info('Label added to task', {
        taskId: input.task_id,
        labelId: input.label_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Label ${input.label_id} added to task "${task.title}"`,
        task,
      };
    } catch (error) {
      logger.error('Failed to add label to task', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to add label to task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Remove a label from a task
   */
  async removeLabel(
    input: RemoveLabelInput,
    userContext: UserContext
  ): Promise<AssignmentToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Remove label from task
      const task = await this.client.delete<VikunjaTask>(
        `/api/v1/tasks/${input.task_id}/labels/${input.label_id}`
      );

      logger.info('Label removed from task', {
        taskId: input.task_id,
        labelId: input.label_id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Label ${input.label_id} removed from task "${task.title}"`,
        task,
      };
    } catch (error) {
      logger.error('Failed to remove label from task', {
        error,
        userId: userContext.userId,
      });
      return {
        success: false,
        message: 'Failed to remove label from task',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * Create a new label
   */
  async createLabel(
    input: CreateLabelInput,
    userContext: UserContext
  ): Promise<AssignmentToolResult> {
    try {
      // Rate limiting check
      await this.rateLimiter.checkLimit(userContext.token);

      // Set auth token
      this.client.setToken(userContext.token);

      // Create label
      const label = await this.client.put<VikunjaLabel>('/api/v1/labels', input);

      logger.info('Label created', {
        labelId: label.id,
        userId: userContext.userId,
      });

      return {
        success: true,
        message: `Label "${label.title}" created successfully with ID ${label.id}`,
        label,
      };
    } catch (error) {
      logger.error('Failed to create label', { error, userId: userContext.userId });
      return {
        success: false,
        message: 'Failed to create label',
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
