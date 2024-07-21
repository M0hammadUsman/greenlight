# Greenlight API

## Overview
This project follows the [Let's Go Further](https://lets-go-further.alexedwards.net/) book by [Alex Edwards](https://www.alexedwards.net/), with some modifications and enhancements.

## Key Differences from the Book

1. **Routing**: Utilizes Go 1.22's standard library for routing, eliminating the need for [httprouter](https://github.com/julienschmidt/httprouter).

2. **Logging**: Implements `slog` with [tint](https://github.com/lmittmann/tint) for enhanced human readability.

3. **Database Connectivity**: Employs [pgx](https://github.com/jackc/pgx) instead of [pq](https://github.com/lib/pq), following current recommendations.

4. **Logging Format**: Uses colorized text output instead of structured JSON logging.

5. **Metrics**: Implements more extensive metrics support, with database-specific metrics adapted for pgx.

6. **Deployment**: Adopts a different approach:
   - Utilizes Amazon EC2 instance as a self-hosted runner
   - Implements GitHub Actions CI/CD pipeline
   - Containerizes the application using Docker

   For more details, refer to the [Dockerfile](https://github.com/M0hammadUsman/greenlight/blob/main/Dockerfile), [compose file](https://github.com/M0hammadUsman/greenlight/blob/main/compose.yml), and [CI/CD pipeline configuration](https://github.com/M0hammadUsman/greenlight/blob/main/.github/workflows/cicd.yml) in the repository.

## Personal Insights

This book is a must-read. Everything you learn from it will be applicable elsewhere in your development future. It follows best practices throughout. 

I've built APIs before moving to Go, primarily using Java Spring Boot. However, after reading this book, I realized I was unaware of several important concepts. This book helped me gain valuable knowledge that I hadn't encountered in my previous development experiences.

Even for beginner level developers with prior API development experience, this book offers insights and best practices that may not be immediately apparent when transitioning to Go. It's an excellent resource for both newcomers to API development and those looking to enhance their skills in Go.
