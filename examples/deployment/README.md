# How to Deploy RINP?

## Introduction

This document provides step-by-step instructions for deploying RINP. Please follow the outlined procedures to compile the Docker image and deploy the various components using the provided YAML files. RINP consists of a client, a reverse proxy, controller, Redis, and an authentication service. It is important to deploy each component in the appropriate environment to ensure proper functionality and security.

## Compiling the Docker Image

To compile the Docker image for RINP, please run following command in root path:

```shell
./init.sh
```

You can also do it step by step:

1. Build a base container image which is useful for testing purposes: `cd examples && make && cd -`
2. Build RINP components using the base container that we just built: `BASE_IMAGE=netutils make container`
3. Prepar a test user: `cp examples/demo.db examples/pb_data/data.db`. You can also change it in `Auth` module.

## Deployment

RINP consists of several components, each with its own Docker Compose file. Follow the instructions below to deploy each component.

### Deploying the Client (client.yml)

The client component can be deployed anywhere. Follow these steps to deploy the RINP client:

1. Retrieve the `client.yml` file from the RINP repository.

2. Deploy the client component using Docker Compose:

   ```shell
   docker-compose -f client.yml up -d
   ```

### Deploying the Reverse Proxy (proxy.yml)

The reverse proxy component should be deployed on a cloud server. Follow these steps to deploy the RINP reverse proxy:

1. Retrieve the `proxy.yml` file from the RINP repository.

2. Deploy the reverse proxy component using Docker Compose on your cloud server:

   ```shell
   docker-compose -f proxy.yml up -d
   ```

### Deploying Other RINP Components (services.yml)

The `services.yml` file is used to deploy other RINP components such as the controller and Redis. These components should not be directly exposed to the public network. Follow these steps to deploy the other RINP components:

1. Retrieve the `services.yml` file from the RINP repository.

2. Deploy the other RINP components using Docker Compose:

   ```shell
   docker-compose -f services.yml up -d
   ```

### Deploying the Authentication Service (auth.yml)

The authentication service provides initial authentication and acts as a special proxy. Follow these steps to deploy the RINP authentication service:

1. Retrieve the `auth.yml` file from the RINP repository.

2. Deploy the authentication service using Docker Compose:

   ```shell
   docker-compose -f auth.yml up -d
   ```

