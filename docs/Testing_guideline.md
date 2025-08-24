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
    - We have role base permission. Please make sure to generate proper access
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
