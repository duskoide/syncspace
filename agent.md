# SyncSpace Edu — Agent Development Guide

**Status:** Early Development / MVP Phase  
**Project Type:** Self-Hosted Educational Productivity Platform  
**Architecture:** Monorepo, Single-Server, Shared Backend Services  
**Primary Goal:** Build a lightweight, web-based educational study management platform.

## Overview

SyncSpace Edu is a self-hosted educational productivity platform designed for efficient workflow and learning management.

Core features:
- task management
- markdown note-taking
- educational resource integration
- lightweight multi-device access

Primary technologies:
- Go backend
- React frontend
- SQLite database
- REST API
- Cloudflare Tunnel deployment

## Assignment Alignment

This project satisfies:
- frontend application
- backend service
- self-managed database
- external/public API integration
- public internet accessibility

## External API Integration

Primary API:
- Wikipedia API

Usage:
- educational topic lookup
- note enrichment
- quick summaries

Example:
1. User searches a study topic
2. Backend requests summary from Wikipedia
3. Summary is inserted into notes

## Architecture

```text
┌─────────────────┐
│ React Frontend  │
└────────┬────────┘
         │ HTTP REST API
┌────────▼────────┐
│ Go Backend Core │
├─────────────────┤
│ Service Layer   │
│ Store Layer     │
│ API Layer       │
└────────┬────────┘
         │
┌────────▼────────┐
│ SQLite Database │
└─────────────────┘

```

## REPOSITORY STRUCTURE
syncspace/
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── api/
│   │   ├── service/
│   │   ├── store/
│   │   ├── models/
│   │   └── config/
│   └── main.go
│
├── frontend/
├── data/
└── agent.md

## Architectural Rules
- All database access must go through the store layer

- No raw SQL outside internal/store

- Frontend communicates exclusively through REST APIs

- SQLite is permanent

- WAL mode must be enabled

# Forbidden:

- ORMs

- GraphQL

- microservices

- Kubernetes

- Docker-first assumptions

## Development Phases
# Phase 1
- SQLite setup

- Task CRUD

- Note CRUD

- REST API foundations

# Phase 2
- Service layer implementation

- Concurrency-safe SQLite usage

- Core API completion

# Phase 3
- Wikipedia API integration

- educational search

- note enrichment

# Phase 4
- React dashboard

- frontend integration with API

# Phase 5
- public deployment

- systemd setup

- Cloudflare Tunnel

## Final Philosophy
SyncSpace Edu should remain:

- lightweight

- maintainable

- web-centric

- self-hostable
