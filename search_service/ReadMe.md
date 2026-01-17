Pet - project: vacancy parser (web version)
stack: golang, gin

done:

- have made design of project
- have made the mechanism of serarching vacancies fron 2 sourses (HH.ru, SuperJob.ru) (parser manager, able to add new sourses of vacancies)
- have made sharded inmemory cache for searching (founded vacancies)
- have made sharded inmemory cache - to get vacancy by ID (use reverse index)
- have made rate limiter for each parser (limiting quantity of requests for each service)
- have made seatch by ID in sharded inmemory cache (reverse index)
- have made "the semaphore" pattern for each parser (concurrency limiting)
- have made circuit breaker for each parser (for blocking unavailable service)
- have made the mechanism of parser creation (factory pattern)
- have made logics of getting data (from .env and from .yml)
- have made global Circuit Breaker - health monitoring [in parsers manager] (**_ needed load testing _**)
- have made queue for parsers manager, have made workers mechanism, workers are handling the jobs in queue
- have made global "semaphore" - limiting quantity of searchings in parsers manager (resourse contron of server)
- have made parsers status manager, heath check control of parsers
- have made the mechanism of update parsers ststus through time interval (time interval - in config)
- have made FIFO queue based on generics, as input - job interface (queue gets jobs interfaces)
- have made unit tests for: rate limiter, circuit breaker, sharded inmemory cache, queue
- have made coucurrent search of vacancies through some quantity of sourses

to be done:

- make layers: service, DTO, REPO
- logging
- handlers
- fix methods of parsers manager
- create data convertors between layers
- create repo layer (working with caches)

Architecture:
user request
↓
┌─────────────────────────────────┐
│ PARSER MANAGER level ← Global control
│ • Global semaphore
│ • Queue
│ • Global circuit breaker
└─────────────────────────────────┘
↓
┌─────────────────────────────────┐
│ PARSER level (HH/SuperJob) ← Individual control
│ • Individual semaphore
│ • Sourse Rate limiter
│ • Sourse Circuit breaker
└─────────────────────────────────┘
↓
External API

repo: https://github.com/AnumSmart/Job_parser.git
