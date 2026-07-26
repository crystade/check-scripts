---
description: Describe when these instructions should be loaded by the agent based on task context
# applyTo: 'Describe when these instructions should be loaded by the agent based on task context' # when provided, instructions will automatically be added to the request context when the pattern matches an attached file
---

## Functions
### 1. Alerting
```rice
alerting.send({
	header: "Monitor check fails", # optional
	message: "A failure has been detected", # required
	footer: "My monitor - My company", # optional
	deduplication: { # optional
		key: MONITOR_ID, # required if "deduplication" exists
		ttl: "5m"        # required if "deduplication" exists
	}
});
```
- The maximum deduplication TTL is 1 day
- The key is always isolated per team

### 2. Incident
#### 2.1. Signal an intent to report incident
- When a check fails, the script can raise a signal that an incident should be reported. To tolerate and avoid false positives, every monitor has a setting "tolerable time window" (between 1m to 7 days)
```rice
incident.signal({
	severity: "minor",
	title: "Monitor %s has been down".format(monitor.name),
	content: "Monitor %s has been down".format(monitor.name)
});
```
- Severity is one `minor, major, critical` per Incident PRD
- Title and Content accept a lambda function so the title and content is rendered 
- There is no incident report if there exists another that is unresolved and created from the same monitor

#### 2.1. Recover an incident intent
```rice
incident.recover();
```
