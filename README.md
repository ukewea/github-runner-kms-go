# GitHub Runner KMS (Go)

This project is a reimplementation of [knatnetwork/github-runner-kms](https://github.com/knatnetwork/github-runner-kms), with HTTPS proxy support. The necessity for this fork arose due to the requirement to contact GitHub from behind a proxy in my work environment. Unlike the original implementation, this version does not include the `/repo/*` endpoint as it was not needed for our purposes.

## Features

- **Proxy Support**: This version supports HTTP and HTTPS proxies, allowing the runner to communicate with GitHub from behind a proxy.
- **Organization PAT Map**: Instead of setting Personal Access Tokens (PATs) through environment variables, this version uses a JSON file, `org-pat-map.json`, to map PATs to GitHub organizations. This change was made to support organization names with hyphens, such as `org-name`.

## Configuration

To use this service, you need to mount the `org-pat-map.json` file to `/app/org-pat-map.json` in the container. This file should contain the PAT tokens for the corresponding GitHub organizations.

### Example `docker-compose.yaml`

```yaml
services:
  runner:
    image: knatnetwork/github-runner:jammy-2.315.0
    restart: always
    environment:
      RUNNER_REGISTER_TO: 'org-name'
      RUNNER_LABELS: 'purpose1,spec2'
      KMS_SERVER_ADDR: 'http://kms:3000'
      ADDITIONAL_FLAGS: '--ephemeral'
      # I bound a CNTLM proxy to the docker0 interface
      http_proxy: "http://172.17.0.1:3128/"
      https_proxy: "http://172.17.0.1:3128/"

      # a must set environment variable for curl to connect directly to kms without proxy
      no_proxy: "kms"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - kms

  kms:
    build: ghcr.io/ukewea/github-runner-kms-go:latest
    restart: always
    environment:
      # I bound a CNTLM proxy to the docker0 interface
      https_proxy: "http://172.17.0.1:3128/"
    volumes:
      - ./org-pat-map.json:/app/org-pat-map.json:ro
```

In this docker-compose.yaml, the runner service is configured to use a GitHub runner image, and the kms service builds from the local github-runner-kms directory. The http_proxy and https_proxy environment variables are set to enable communication through the proxy.


### Example `org-pat-map.json`

The `org-pat-map.json` file should follow the below format, mapping your GitHub organization names to their corresponding PAT tokens:

```json
{
    "org-name": "PAT token"
}
```

Replace `org-name` with your actual GitHub organization name and `PAT token` with the personal access token generated for that organization.

## Usage
To start using this implementation:

1. Ensure you have Docker and Docker Compose installed.
2. Create an `org-pat-map.json` file with the necessary PAT tokens and organization mappings.
3. Use the provided `docker-compose.yaml` as a template, adjusting as necessary for your environment. For detailed configuration options and usage instructions, refer to the documentation in [knatnetwork/github-runner](https://github.com/knatnetwork/github-runner)
4. Run `docker compose up` to start the services.

Note that the environment variables, network settings, and volume mounts should be adapted to fit your specific requirements and environment.
