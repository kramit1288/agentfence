# Tests

The repository keeps test layers separate so critical path behavior can expand without mixing concerns:

- `tests/unit`: package-focused behavior checks
- `tests/integration`: component and storage integration coverage
- `tests/e2e`: gateway-level end-to-end scenarios

No synthetic tests are added in the initial scaffold. Real tests should be introduced alongside real behavior.
