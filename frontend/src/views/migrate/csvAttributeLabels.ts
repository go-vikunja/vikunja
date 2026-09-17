// Keys are the snake_case attribute names the CSV migration API uses.
export const CSV_ATTRIBUTE_LABEL_KEYS: Record<string, string> = {
	title: 'task.attributes.title',
	description: 'task.attributes.description',
	due_date: 'task.attributes.dueDate',
	start_date: 'task.attributes.startDate',
	end_date: 'task.attributes.endDate',
	done: 'task.attributes.done',
	priority: 'task.attributes.priority',
	labels: 'task.attributes.labels',
	reminder: 'task.attributes.reminders',
	project: 'task.attributes.project',
	ignore: 'migrate.csv.ignore',
}
