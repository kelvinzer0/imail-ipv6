<p align="center">
  <img alt="logo" src="https://avatars2.githubusercontent.com/u/4169529?v=3&s=200" height="140" />
  <h3 align="center">imail</h3>
  <p align="center">imail is an easy-to-set-up self-service email server.</p>
</p>


---
## Project Vision

The imail project aims to build a simple and stable email service in the easiest way possible. Developed in Go, imail can be distributed as a standalone binary and supports all platforms supported by Go, including Linux, macOS, Windows, and ARM platforms.

- Supports multi-domain management.
- Email draft functionality.
- Email search functionality.
- Rspamd spam filtering support.
- Hook script support.

[![Go](https://github.com/kelvinzer0/imail/actions/workflows/go.yml/badge.svg)](https://github.com/kelvinzer0/imail/actions/workflows/go.yml)
[![CodeQL](https://github.com/kelvinzer0/imail/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/kelvinzer0/imail/actions/workflows/codeql-analysis.yml)
[![Codecov](https://codecov.io/gh/kelvinzer0/imail/branch/master/graph/badge.svg?token=MJ2HL6HFLR)](https://codecov.io/gh/kelvinzer0/imail)

## Screenshots

[![main](/screenshot/main.png)](/screenshot/main.png)


## Version Details

- 0.0.18

```
* Added password modification function for administrators.
* Optimized log display.
* initd changed to systemd.
* Fixed the phenomenon of being unable to log in during initialization.
* SSL function optimization.
* Optimized some prompts.
```

## Build Dependencies

```
go install -a -v github.com/go-bindata/go-bindata/...@latest
```

## Installation

To install imail, you can build it from source. Make sure you have Go installed (version 1.16 or higher recommended).

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/kelvinzer0/imail-ipv6.git
    cd imail-ipv6
    ```
2.  **Build the project:**
    ```bash
    go build ./...
    ```
    This will create an executable file in your current directory (e.g., `imail.exe` on Windows, `imail` on Linux/macOS).
3.  **Run the application:**
    ```bash
    ./imail service
    ```
    (Replace `./imail` with `imail.exe` on Windows)

## Uninstallation

To uninstall imail, simply remove the cloned repository directory and any generated executable files.

```bash
rm -rf imail-ipv6
# If you installed the binary to your PATH, you might need to remove it manually.
# For example: rm /usr/local/bin/imail
```

## Swagger API Documentation

The backend API documentation is automatically generated using Swagger.

1.  **Generate Swagger documentation:**
    Ensure you have `swag` installed:
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
    Then, generate the documentation from the project root:
    ```bash
    swag init -generalInfo imail.go
    ```
    This will create `docs` directory containing `docs.go`, `swagger.json`, and `swagger.yaml`.

2.  **Access Swagger UI:**
    Once the `imail` application is running, you can access the Swagger UI in your web browser at:
    `http://localhost:<port>/swagger/index.html`
    (Replace `<port>` with the actual port your imail application is running on, e.g., 8080).

## Wiki

- https://github.com/kelvinzer0/imail/wiki

## Contributors

[![](https://contrib.rocks/image?repo=kelvinzer0/imail)](https://github.com/kelvinzer0/imail/graphs/contributors)


## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/kelvinzer0/imail/blob/main/LICENSE) file for full details.