# Contributing to ipmap

Thank you for your interest in contributing to ipmap! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs
- Check if the bug has already been reported in [Issues](https://github.com/sercanarga/ipmap/issues)
- Provide a clear description of the bug
- Include steps to reproduce the issue
- Attach relevant logs or error messages

### Suggesting Enhancements
- Check if the enhancement has already been suggested
- Provide a clear description of the proposed feature
- Explain the use case and benefits

### Pull Requests
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass (`go test ./... -v`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/sercanarga/ipmap.git
cd ipmap

# Install dependencies
go mod download

# Build the project
go build -o ipmap .

# Run tests
go test ./... -v

# Run with verbose output
./ipmap -asn AS13335 -v
```

## Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and concise
- Run `go fmt` before committing

## Testing

- Write tests for new features
- Ensure all existing tests pass
- Aim for good test coverage
- Test edge cases and error conditions

## License

By contributing to ipmap, you agree that your contributions will be licensed under the MIT License.
