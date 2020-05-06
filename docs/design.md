# Design

## Scanner Flow

Scanner starts up with configuration.

## Browsers

A browser is an implementation of a gcd.Tab. The browser pool handles acquiring new browsers and returning old ones. The pool gets browsers from the leaser service which handles starting new ones and closing old ones.

## Storage

Pretty much every data type should be stored in our custom DB built on badger

- Crawl Graph
- Attack Graph / Plugin Work Graph
- Store req/resp in seperate badger nodes with requestID as key? keep graphs light?

## Plugins

Should support running external commands to get easy wins for coverage (TLS testing etc)
Should be configurable for types:

- run once
- run once per path
- once per page
- run only on X mimetypes
- run only on X injection point types
- need ability to send direct requests for certain plugin types (might have to rewrite devtool methods/inject capabilities)
- should plugins have dependencies (on other plugins)?

Active Plugins are directly handed the navigation actions done by a crawler and re-execute them with a reference to the browser, this way they have full control over the entire navigation path and can know when to attack without requiring tracking which requests/responses to attack. Effectively making plugins first class citizens.

Passive Plugins register listeners for the types of data they want (storage events, network events, cookie events) and a passive manager filters out duplicates then dispatches new events to them to process.

## Crawling

Crawling data is stored as a graph with properities of visited or not. This will allow for using the graph as a queue for edges
to traverse to find new actions/pages.

- Page/Action should have a status property for a state machine, with timers (maybe store a timestamp for SM) for the event that actions fail and a fail count to 'give up'
  - Potential
  - Queued
  - Visited

## Authentication and Session Management

TODO: maybe support loading selenium/injecting selenium into browserker so we get selenium capabilities for scripting?
Supporting things like JWT should be easy (we can inject whatever we want into browser processes)

## Attacking

TODO: What should plugins _get_? A list of injection points? A page? A browser? Register for specific events? Needs access to responses.
Needs ability to read response for _their_ injected request.

### Parsing Requests

Parse methods, request headers, x-www-form-urlencoded, url's/path queries and fragments, json and XML by hand using injast (parses data into an AST).

### Injection Points

Each k/v for parsed types should be an injection point. Injection Manager should handle re-encoding (similar to how astutil works in Go)

## Reporting

Report manager should be available to plugins, plugins can report their specific checks with evidence.
