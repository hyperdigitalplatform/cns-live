# RTA CCTV System - Testing

This directory contains test infrastructure for the RTA CCTV system.

## Test Categories

### 1. **Unit Tests**
- Located within each service directory (`services/*/internal/`)
- Run with `go test ./...` in each service
- Coverage target: >80%

### 2. **Integration Tests** (`integration/`)
- Test service-to-service communication
- Use Testcontainers for dependencies
- Run with Docker Compose test profile

### 3. **End-to-End Tests** (`e2e/`)
- Test complete user workflows
- Use Playwright or Cypress
- Run against deployed system

### 4. **Load Tests** (`load/`)
- Performance and stress testing
- Use k6 or JMeter
- Test scalability and resource usage

## Running Tests

### Unit Tests
```bash
# Run all unit tests
cd services/<service-name>
go test ./... -v -cover

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Start test environment
cd tests/integration
docker-compose up -d

# Run integration tests
go test ./... -v

# Cleanup
docker-compose down -v
```

### E2E Tests
```bash
# Install dependencies
cd tests/e2e
npm install

# Run tests
npm test

# Run with UI
npm run test:ui
```

### Load Tests
```bash
# Install k6
# https://k6.io/docs/getting-started/installation/

# Run load test
cd tests/load
k6 run stream-load-test.js
```

## Test Status

| Test Type | Status | Coverage | Priority |
|-----------|--------|----------|----------|
| Unit Tests | ⏸️ TODO | 0% | High |
| Integration Tests | ⏸️ TODO | 0% | High |
| E2E Tests | ⏸️ TODO | 0% | Medium |
| Load Tests | ⏸️ TODO | 0% | Medium |

## TODO - Testing Implementation

### Phase 1: Unit Tests (Week 1-2)
- [ ] VMS Service unit tests
- [ ] Stream Counter unit tests
- [ ] Storage Service unit tests
- [ ] Recording Service unit tests
- [ ] Metadata Service unit tests
- [ ] Playback Service unit tests
- [ ] Go API unit tests

### Phase 2: Integration Tests (Week 3-4)
- [ ] Service-to-service communication tests
- [ ] Database integration tests
- [ ] MinIO integration tests
- [ ] Valkey integration tests
- [ ] LiveKit integration tests

### Phase 3: E2E Tests (Week 5-6)
- [ ] User authentication flow
- [ ] Live stream viewing
- [ ] PTZ control
- [ ] Playback video
- [ ] Search and filter
- [ ] Incident creation

### Phase 4: Load Tests (Week 7-8)
- [ ] Concurrent stream reservations
- [ ] API endpoint load
- [ ] Database query performance
- [ ] Cache performance
- [ ] Playback concurrency

## Test Data

### Fixtures
- Located in `fixtures/` directory
- Sample camera configurations
- Mock video segments
- Test user data

### Test Cameras
- Use test RTSP streams
- RTSP test server: `rtsp://test.example.com/stream`
- Or use MediaMTX test streams

## CI/CD Integration

### GitHub Actions (Example)
```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: |
          cd services
          for service in */; do
            cd $service
            go test ./... -v -cover
            cd ..
          done

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Start services
        run: docker-compose up -d
      - name: Run integration tests
        run: |
          cd tests/integration
          go test ./... -v
      - name: Cleanup
        run: docker-compose down -v
```

## Contributing

When adding new features:
1. Write unit tests first (TDD)
2. Ensure >80% coverage
3. Add integration tests for service interactions
4. Update E2E tests for user-facing changes
5. Run all tests before submitting PR

## Additional Resources
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testcontainers](https://www.testcontainers.org/)
- [Playwright](https://playwright.dev/)
- [k6 Documentation](https://k6.io/docs/)
