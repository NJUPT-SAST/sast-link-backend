# SAST Link

![SAST Link Logo](https://aliyun.sastimg.mxte.cc/images/2023/07/02/footera9663bd5ff4b2bad.png)

Logo designed by [SAST](https://sast.fun/), created by [Maxtune Lee](https://github.com/MaxtuneLee).

[![Go Report Card](https://goreportcard.com/badge/github.com/NJUPT-SAST/sast-link-backend)](https://goreportcard.com/report/github.com/NJUPT-SAST/sast-link-backend)
[![License](https://img.shields.io/badge/license-AGPLv3-blue.svg)](https://choosealicense.com/licenses/agpl-3.0/)

SAST Link is a comprehensive personnel management system and OAuth designed to provide a secure and efficient way to manage and authorize access to your applications and services. 

Product design in Figma: [SAST Link](https://www.figma.com/file/IUIoRll3ieYFzJSfJPelDu/sast-link?node-id=0-1&t=rtc1sJfjJ0aTDAkp-0), designed by [Maxtune Lee](https://github.com/MaxtuneLee)

This repository contains the backend code for SAST Link. If you're interested in the frontend, please visit [SAST Link frontend](https://github.com/NJUPT-SAST/sast-link).

SAST Link backend is built with Go and PostgreSQL, and use gin as the web framework.

> [!WARNING]
> This repo is under active development! Formats, schemas, and APIs are subject to rapid and backward incompatible changes!

## Getting Started

### Pre-requisites

- Go
- PostgreSQL
- Redis
- Email Account (SMTP)
- Tencent COS (For file storage)
- Oauth2.0 Provider (e.g. GitHub, Feishu)

Create PostgreSQL database and tables by running the SQL scripts in `sql/` directory.

### Quick Start

First, create a configuration file based on `src/config/example.toml`. Ensure that you provide appropriate configurations for your environment.

```bash
git clone https://github.com/NJUPT-SAST/sast-link-backend.git && cd sast-link-backend

docker-compose up --detach
```

```
[+] Running 3/3
 ✔ Container sast-link-backend-redis-1              Healthy                                          0.0s
 ✔ Container sast-link-backend-postgres-1           Heal...                                          0.0s
 ✔ Container sast-link-backend-sast-link-backend-1  Started                                          0.0s
```

These commands will build and start services listed in the compose file:

- configuration and start postgreSQL
- configuration and start redis
- start SAST Link

The postgreSQL and redis services are required for the SAST Link service to run.

The `.env` file contains the environment variables for the SAST Link service:

```
POSTGRES_DB=sastlink
POSTGRES_USER=sastlink
POSTGRES_PASSWORD=sastlink
REDIS_PASSWORD=sastlink
CONFIG_FILE=dev-example
```

The `CONFIG_FILE` environment variable is used to specify the configuration file for the SAST Link service.
The `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `REDIS_PASSWORD` environment variables are used to initialize the postgreSQL and redis services.

## Development

### API Documentation

The API documentation is available at [wiki](https://github.com/NJUPT-SAST/sast-link-backend/wiki/Api-Doc)

### Database Schema

The database schema is available at [wiki](https://github.com/NJUPT-SAST/sast-link-backend/wiki/Project-Structure#sql)

### Code Workflow Explanation

The code workflow is available at [wiki](https://github.com/NJUPT-SAST/sast-link-backend/wiki/General)

## Roadmap

Goals and Vision for SAST Link (SAST OAuth and SAST Profile):

**SAST OAuth:**

SAST OAuth serves as a unified identity authentication system for SAST, facilitating login across multiple SAST applications.

Example:

- Simplifies login processes for SAST members across various projects, such as the FreshCup competition.
- Enables seamless login via SAST credentials without the need for separate accounts for each project.
- Allows SAST lecturers to access and manage the FreshCup competition system for tasks like grading via SAST login.
- Offers multiple login options including SAST Feishu, PassKey, QQ, Github, etc., providing users with convenience and flexibility.
- Implements additional security measures like F2A and security keys to enhance account security.

In login process, users can choose to log in in multiple ways: SAST Feishu, PassKey, QQ, Github, etc. As long as they have been bound in advance, they can use third-party login, which is convenient and fast. They can also use F2A, security keys, and other methods to enhance account security.

**SAST Profile:**

SAST Profile acts as a centralized user profile system for managing user information and settings within SAST applications.

Features:

- Records basic user information such as SAST membership status, current position, department, group affiliation, etc.
- Tracks user activities within SAST, including competition results, awards, and permissions across various applications.
- Provides users with the ability to customize and share their profile page, allowing them to control the visibility of their information.

**Current status**:

- [x] User Management (Basic)
- [x] SAST OAuth (Basic)
- [x] File Storage (Tencent COS)
- [x] SAST Profile (Basic)
- [ ] SAST Link management
- [ ] Third-party OAuth (Github and Feishu now can be used in backend, but not fully implemente)
- [ ] CI/CD, Docker, and Kubernetes support

## Contributing

Pull requests and any feedback are welcome. For major changes, please open an issue first
to discuss what you would like to change.

### Contributors

<a href="https://github.com/NJUPT-SAST/sast-link-backend/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=NJUPT-SAST/sast-link-backend" />
</a>

## License

[AGPLv3 ](https://choosealicense.com/licenses/agpl-3.0/)
