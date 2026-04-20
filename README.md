# flam-in-test

A test framework built on top of [flam-in-go](https://github.com/cjdias/flam-in-go), providing a structured approach to testing Go applications using the Arrange-Act-Assert pattern.

## Features

- **Arrange-Act-Assert Pattern**: Structured test execution flow
- **Dependency Injection Support**: Seamless integration with flam-in-go's container
- **Database Testing**: Built-in support for database transactions with automatic rollback
- **Redis Testing**: Redis connection management for integration tests
- **PubSub Testing**: Message publishing and subscription verification
- **Configuration Management**: Load and manage test configurations from YAML files
- **Flexible Assertions**: Built-in assertion helpers for common scenarios
- **Logging**: Optional logging for test execution steps

## Installation

```bash
go get github.com/cjdias/flam-in-test
```

## Usage

### Basic Example

```go
package myapp_test

import (
    "testing"
    "github.com/cjdias/flam-in-go"
    "github.com/cjdias/flam-in-test"
)

func TestMyFeature(t *testing.T) {
    app := flam.NewApplication()
    runner, err := test.NewRunner(app, t)
    if err != nil {
        t.Fatal(err)
    }
    defer runner.Close()

    runner.
        WithLogger(test.NewLogger()).
        WithConfig("config/test.yaml", 1).
        WithArrangeDatabase("default").
        WithArrange("setup user", func(db *gorm.DB) {
            db.Create(&User{Name: "test"})
        }).
        WithAct(func(service MyService) error {
            return service.DoSomething()
        }).
        WithAssertNoError(&err).
        WithTeardownDatabase().
        Run()
}
```

### Database Testing

```go
runner.
    WithArrangeDatabase("default").
    WithAct(func(repo UserRepository) error {
        return repo.Create(&User{Name: "John"})
    }).
    WithAssert("user created", func(db *gorm.DB) {
        var count int64
        db.Model(&User{}).Count(&count)
        assert.Equal(t, int64(1), count)
    }).
    WithTeardownDatabase().
    Run()
```

### Redis Testing

```go
runner.
    WithArrangeRedis("default").
    WithAct(func(cache CacheService) error {
        return cache.Set("key", "value")
    }).
    WithAssert("value set", func(client redis.Client) {
        val, err := client.Get("key").Result()
        assert.NoError(t, err)
        assert.Equal(t, "value", val)
    }).
    Run()
```

### PubSub Testing

```go
runner.
    WithArrangePubSub().
    WithArrangeSignals([]string{"events"}).
    WithAct(func(publisher Publisher) error {
        return publisher.Publish("events", "message")
    }).
    WithAssertPublished([]string{"events"}).
    Run()
```

## API Reference

### Runner Methods

- `NewRunner(app flam.Application, t *testing.T) (*Runner, error)` - Create a new test runner
- `WithLogger(logger Logger) *Runner` - Set a logger for test execution
- `WithConfig(file string, priority int) *Runner` - Add a configuration file
- `WithProcess(id string) *Runner` - Activate a specific process
- `WithArrange(id string, function any) *Runner` - Add an arrange step
- `WithArrangePubSub() *Runner` - Setup pubsub for testing
- `WithArrangeSignals(channels []string) *Runner` - Setup signal subscriptions
- `WithArrangeDatabase(connectionId string) *Runner` - Setup database transaction
- `WithArrangeRedis(connectionId string) *Runner` - Setup redis connection
- `WithAct(function any) *Runner` - Set the act step
- `WithAssert(id string, function any) *Runner` - Add an assertion
- `WithAssertNoError(err *error) *Runner` - Assert no error occurred
- `WithAssertError(id string, err *error, expected error, msg ...string) *Runner` - Assert specific error
- `WithAssertTotal(expected int64, total *int64) *Runner` - Assert record count
- `WithAssertBoolResponse(res *bool, expected bool) *Runner` - Assert boolean response
- `WithAssertPublished(calls []string) *Runner` - Assert published messages
- `WithTeardown(id string, function any) *Runner` - Add a teardown step
- `WithTeardownDatabase() *Runner` - Add database rollback teardown
- `Run() error` - Execute the test
- `Close()` - Cleanup resources

## Development

### Running Tests

```bash
task quality:tests
```

### Running Tests with Coverage

```bash
task quality:tests-coverage
```

### Linting

```bash
task quality:lint
```

### Formatting

```bash
task quality:format
```

## License

See LICENSE file for details.
