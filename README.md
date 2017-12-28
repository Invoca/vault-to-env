# vault-to-env

`vault-to-env` originates from the common scneario that applications need secrets to run. Especially in containers, we desired a solution that would query Vault and set the resulting values to ENV variables.

At the very least, this standardizes the ENTRYPOINT querying of Vault and allows the remaining logic to implement the values as necessary in the container at run time.

## Getting Started

The application is available on DockerHub at `invocaops/vault-to-env`. With that in mind, the easiest way to retrieve and use the application is with a [multi-stage build](https://docs.docker.com/engine/userguide/eng-image/multistage-build/).

### Dockerfile

```Dockerfile
FROM invocaops/vault-to-env as vte

FROM your-image
COPY --from=vte /vault-to-env /usr/local/bin/
COPY docker-entrypoint.sh /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
```

### docker-entrypoint.sh

```bash
#!/bin/bash

eval `vault-to-env -url $VAULT_ADDR \
    -token $VAULT_TOKEN \
    -path secret/supersecret \
    -eks SECRET_USER=user \
    -eks SECRET_PASSWORD=password`

...
```

### docker run

```bash
docker run -e VAULT_ADDR=https://localhost:8200 -e VAULT_TOKEN=supersecrettoken your-image
```
