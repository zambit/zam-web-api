# Inter Services Communication

This document describes communication model used by this project in micro-services environment.

## Model

The communication model very close to the Actors Model, except some assumptions.

### Communication message attributes

Despite of underlying broker implementation, all messages have 4 attributes:

1. Resource - this is address of specific actor in the system
2. Action - required action (like rpc method) or event name (there is assumption that all events actions appended with "_event")
3. ID - optional message identifier, strongly required when method call expects response from specific resource.
4. Data - parameters passed into method call or payload associated with an event
