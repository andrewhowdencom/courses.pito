# Delivery Service

Throughout this course, this application will form a "persistent example" that we can return to to understand the
concepts of observability.

⚠️ This README has been written before the actual application, in a
[documentation first](https://dev.to/eminetto/document-first-48dh) approach. At the time of writing, it is not complete.

## Architecture

The application is a RESTful API that you can query to fetch delivery options from different (fake) third-party
providers. I mocked the third-party providers and made them generate random results and, occasionally, unexpected
failures. The (nonsensical) providers include:

* **svx**: Stock Variant Express
* **mmc**: Million Mile Company
* **hid**: High Inertia Delivery

The application publishes the results of the successful queries to an event stream to mock an "analytics" workflow.

## Observability

The application is deliberately written entirely without any telemetric output of any kind. It is an exercise during
the course for learners to implement the various types of telemetry and make the application "observable."

## Development
### Requirements

* The [Go programming language](https://go.dev/learn/)

### Build

You can build the application via:

```bash
go build
```

Once built, you can run the application:

```bash
./delivery-service
```

When running, you can open another terminal window and use standard tools (e.g., the browser or `curl`) to view the
result.

```bash
curl 'localhost:9093/delivery-options?width=200&height=35&depth=150&weight=2500'

# {
#    ... (The Result)
# }
```

### Test

You can also test the application via:

```bash
go test
```

By default, the tests include tests to validate that the application produces the required observability. I designed
this so learners can submit pull requests, making the application "observable," and use the test suite to validate
whether their changes are successful.

To test only the application logic, use:

```bash
go test -run '!Observe'
```