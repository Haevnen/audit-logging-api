## How to test
- I created a seed record for test tenant. Please use this id to generate access
  token.
    ```bash
    Name      |Value                               |
    ----------+------------------------------------+
    id        |c2e217a3-fe51-444c-a41c-6956c4c81d51|
    name      |Test Company                        |
    ```
- **NOTE**:
    - We have a reference between tenants and logs table. Please ensure creating
      tenant record before adding new log entry.
    - We have role base permission. Make sure to generate proper access
      token for each role.

- POST /auth/token
    ```bash
    curl --location 'localhost:38081/api/v1/auth/token' \
    --header 'Content-Type: application/json' \
    --data '{
    "role": "user",
    "user_id": "123e4567",
    "tenant_id": "c2e217a3-fe51-444c-a41c-6956c4c81d51"
    }'
    ```
- Install `websocat` to verify `WS /api/v1/logs/stream` API
    ```bash
    brew install websocat
    websocat "ws://localhost:38081/api/v1/logs/stream?tenant_id=c2e217a3-fe51-444c-a41c-6956c4c81d51" \                1 ✘  at 10:23:15 
    -H="Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiIxMjNlNDU2NyIsIlRlbmFudElEIjoiYzJlMjE3YTMtZmU1MS00NDRjLWE0MWMtNjk1NmM0YzgxZDUxIiwiUm9sZSI6ImFkbWluIiwiZXhwIjoxNzU2MDA4ODQxLCJpYXQiOjE3NTYwMDUyNDF9.-T1mG2rh5n6LoS1Kz7WGWrTqpT9AbEbqlfol5-x37OM"
    ```
