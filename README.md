# telemetry-api

A GraphQL API for telemetry data from dimo devices.

Run `make` to see some helpful sub-commands:

```
% make

Specify a subcommand:

  `build`               Compile the Go code and output the binary to the `bin/` directory.
  `run`                 Build the project (if not already built) and run the binary.
  `clean`               Remove the `bin/` directory.
  `install`             Build the project and copy the binary to the `INSTALL_DIR`.
  `test`                Run tests.
  `lint`                Lint the project.
  `format`              Format the project.
  `docker`              Build a Docker image of the project.
  `gqlgen`              Generate gqlgen code.
  `gql-model`           Generate gqlgen data model.
  `tools`               Install `golangci-lint`, `gqlgen`, and `model-garage`.
```

## License

[Apache 2.0](LICENSE)
