import { VikunjaClient } from './client.js';
import { VikunjaLabel } from './types.js';

/**
 * Input for creating a label
 */
export interface CreateLabelInput {
  title: string;
  description?: string;
  hex_color?: string;
}

/**
 * Label API methods
 */
export class LabelAPI {
  constructor(private client: VikunjaClient) {}

  /**
   * Get all labels
   * @returns List of labels
   */
  async getLabels(): Promise<VikunjaLabel[]> {
    return this.client.get<VikunjaLabel[]>('/api/v1/labels');
  }

  /**
   * Create a new label
   * @param data - Label data
   * @returns Created label
   */
  async createLabel(data: CreateLabelInput): Promise<VikunjaLabel> {
    return this.client.put<VikunjaLabel>('/api/v1/labels', data);
  }

  /**
   * Add a label to a task
   * @param taskId - Task ID
   * @param labelId - Label ID
   */
  async addLabelToTask(taskId: number, labelId: number): Promise<void> {
    await this.client.put(`/api/v1/tasks/${taskId}/labels`, { label_id: labelId });
  }

  /**
   * Remove a label from a task
   * @param taskId - Task ID
   * @param labelId - Label ID
   */
  async removeLabelFromTask(taskId: number, labelId: number): Promise<void> {
    await this.client.delete(`/api/v1/tasks/${taskId}/labels/${labelId}`);
  }
}
