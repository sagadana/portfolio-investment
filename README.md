# Portfolio Investment Management Service

Manage Portfolio Investments

## Prerequisites

- Docker

## Set up

Copy `.env.example` into `.env`

## Test

Run `docker compose up tester`

_NOTE: It might take a while to set up before executing test_

## Funds Allocation Strategy

1. First allocate funds to 'one-time' plan portfolios till planned amount is met
2. Then allocate funds to 'monthly' plan portfolios till planned amount is met
3. If both 'one-time' & 'monthly' planned amount are met, distribute funds equally to both
