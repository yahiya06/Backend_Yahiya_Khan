# Threat Alert API

This is my solution for the Bugsmirror backend intern task. It is a small Go
server that keeps everything in memory. The SDK sends threat alerts to it, and
people from a company can log in, look at their alerts and resolve them.

## How to run

You need Go 1.22 or newer.

go run .

It starts on port 8080, or on whatever is set in the PORT variable.

To run the tests:

go test ./...
go test -race ./...

The -race one did not work on my laptop. I am on Windows without a C compiler,
so Go says "-race requires cgo". The tests pass with plain go test. I know you
will run -race on your side.

## Project structure

go.mod - module name and go 1.22, no third party packages
main.go - reads PORT and starts the server
server.go - routes and small helpers (json, errors, random hex, time)
models.go - the structs for company, user, app and alert
store.go - all the data in maps, one mutex, and every operation on the data
auth.go - checks the bearer token, the admin role and the X-App-Key
handlers.go - one handler per route plus the input checks
server_test.go - the three tests from the paper
README.md - this file

## What works and what does not

All the routes work and I went through the 22 self-check steps in Postman.
Step 22 (50 at once) I checked with the concurrency test, not by hand.
Not done: I did not do the bonus items. Data is lost when the server restarts
because there is no database.

## How AI helped

- Tools I used: Claude
- Two prompts I used (paste them as you typed them):
  1. Build a Threat Alert REST API in Go using standard library only. Follow the assignment requirements, use in-memory storage, implement authentication, authorization, alert deduplication and concurrency safety. Keep the code simple and modular.
  
  2. Build a Threat Alert REST API in Go using standard library only. Follow the assignment requirements, use in-memory storage, implement authentication, authorization, alert deduplication and concurrency safety. Keep the code simple and modular.

- One thing the AI got wrong, and how I found and fixed it: Claude told me that sending severity as a string ("5" with quotes) would give 400. I tested it in Postman and it returned 201, so the string was acceptedI fixed it by reading severity as json.RawMessage and returning 400 if it starts with a quote. I tested again and got 400.

- One part I wrote or changed myself without AI: I changed some of the error messages in handlers.go and auth.go to make them shorter and clearer.

## Questions

1. Two identical alerts arrive at the same moment. Point to the lines in your
code that make sure only one alert is created.

In ReportAlert in store.go, the first two lines take the lock
(s.mu.Lock and defer s.mu.Unlock), lines 113 and 114 . Looking for an open
alert and creating a new one both happen while I hold that lock, so only one
request can be in there at a time. The second request then finds the alert
the first one made and just adds 1 to count.

2. Why does a user from another company get 404 and not 403?

Because 403 would tell them the alert is real but belongs to someone else.
With 404 a missing alert and another company's alert look the same, so they
can't use it to find out what exists.