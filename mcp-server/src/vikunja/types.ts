/**
 * Vikunja API type definitions
 */

export interface VikunjaUser {
  id: number;
  username: string;
  email: string;
  name: string;
  created: string;
  updated: string;
}

export interface VikunjaProject {
  id: number;
  title: string;
  description: string;
  owner: VikunjaUser;
  created: string;
  updated: string;
  is_archived: boolean;
  hex_color: string;
  parent_project_id: number;
}

export interface VikunjaTask {
  id: number;
  title: string;
  description: string;
  done: boolean;
  done_at: string | null;
  due_date: string | null;
  priority: number;
  labels: VikunjaLabel[];
  assignees: VikunjaUser[];
  project_id: number;
  created: string;
  updated: string;
  created_by: VikunjaUser;
}

export interface VikunjaLabel {
  id: number;
  title: string;
  description: string;
  hex_color: string;
  created: string;
  updated: string;
}

export interface VikunjaComment {
  id: number;
  comment: string;
  author: VikunjaUser;
  task_id: number;
  created: string;
  updated: string;
}

export interface VikunjaTeam {
  id: number;
  name: string;
  description: string;
  members: VikunjaUser[];
  created: string;
  updated: string;
}

export interface VikunjaBucket {
  id: number;
  title: string;
  project_id: number;
  limit: number;
  position: number;
  created: string;
  updated: string;
}
