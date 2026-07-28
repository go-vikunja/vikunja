# Urgency

To prioritize tasks in an order that feels natural, let's sort all of them by multiple properties, each with unique significance to you.

You could sort by one or all of these properties:

- Proximity to due date
- Priority
- Progress
- Matches filter

And combine the property with individual weights, like `1` for Priority and `2` for Progress.

## How does it work?

Vikunja uses a weighted sort by "scoring" each task on each selected property, scaling each subscore by your weights, adding them into a total task score, and sorting all tasks by their total.

> Score = Property1 * Weight1 + Property2 * Weight2 + ...

To make the weights easier to understand and use across different properties, they must scale the same way. We constrain the numeric values to a uniform scale using 0.00 to 1.00 (0-100%):

- Proximity to due date
	+ linear:
		+ 0 for more than 2 weeks away (< -14)
		+ 0.20 for day -14
		+ 0.24 for day -13
		+ etc, up to 1 for 7 days past due.
	+ exponential:
		+ y(x) = e^( (x−7) /5)
			+ In PostgresQL: least(1, exp( (extract(day from due_date - localtimestamp)-7) / 5.0 ))
			+ 7 shifts e^x over to day 7 = 1.0
			+ 5 adjusts the growth rate. modest growth 2 weeks beforehand to 0.25, then big growth to 1 over 7 days
		+ y(x) = 2^( (x−7) /3)
			+ 2^x with growth rate 3 is roughly equal to e^x with 5.
- Priority - enum of `Low` (1) to `DO NOW` (5) mapped linearly to 0-1
- Progress - 0-100% as 0-1
- Matches filter - 1 for match, 0 otherwise
	* `project = Inbox`
	* `labels = foo`

So one example might be:

- PastDue = true; Priority = 1; Progress = 50%
- Score = PastDue * 1000 + Priority * 100 + Progress * 10
- Score = 1 * 1000 + 0.2 * 100 + 0.5 * 10
- Score = 1025

TODO: mention weight normalization before its returned in API?
TODO: mention when urgency scores are calculated and returned. i.e. only single project ID scenarios

By default, weights are set to:

- Due Date = 100
- Percent Done = 100
- Priority = 10
