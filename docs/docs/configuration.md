# Configuration

Compass binary contains both the CLI client and the server. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, elastic search brokers, log severity, etc. while CLI client configuration only has configuration about which server to connect. 
## Server Setup

There are several approaches to setup Compass Server

1. [Using the CLI](#using-the-cli)
1. [Using the Docker](#using-the-docker)
2. [Using the Helm Chart](#using-the-helm-chart)

#### General pre-requisites

- PostgreSQL (version 13 or above)
- ElasticSearch v7 (optional)

## Using the CLI
### Using config file

Create a config file with the following command 

```sh
$ compass config init
```

alternatively you can [use `--config` flag](#using---config-flag) to customize to config file location.

You can also [use environment variables](#using-environment-variable) to provide the server configuration. 

Setup up the Postgres database, and ElasticSearch instance and provide the details as shown in the example below.

> If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample server configuration yaml:

```yaml title="compass.yaml"
log_level: info                                 # debug|info|warning|error|fatal|trace|panic - default: info

elasticsearch:
    brokers: http://localhost:9200              #required
    username:                                   
    password:                                   

db:
    host: localhost                             #required
    port: 5432                                  #required
    name: compass                               #required
    user: compass                               #required
    password: compass_password                  #required
    sslmode: disable                            #optional

service:
    host: localhost                             #required
    port: 8080                                  #required    
    identity:                                   
        headerkey_email: Compass-User-Email     #optional
        provider_default_name: shield           #optional
    grpc:
        port: 8081                              #required
        max_send_msg_size: 33554432     
        max_recv_msg_size: 33554432
```
### Using environment variable

All the server configurations can be passed as environment variables using underscore _ as the delimiter between nested keys. 

See [configuration reference](./reference/configuration.md) for the list of all the configuration keys.

```sh title=".env"
LOG_LEVEL=info
ELASTICSEARCH_BROKERS=http://localhost:9200
ELASTICSEARCH_USERNAME=
ELASTICSEARCH_PASSWORD=
DB_HOST=localhost
DB_PORT=5432
DB_NAME=compass
DB_USER=compass
DB_PASSWORD=compass_password
DB_SSLMODE=disable
SERVICE_HOST=localhost
SERVICE_PORT=8080
SERVICE_IDENTITY_HEADERKEY_EMAIL=Compass-User-Email
SERVICE_IDENTITY_PROVIDER_DEFAULT_NAME=shield
SERVICE_GRPC_PORT=8081
SERVICE_GRPC_MAX_SEND_MSG_SIZE=33554432
SERVICE_GRPC_MAX_RECV_MSG_SIZE=33554432
```

Set the env variable using export
```bash
$ export DB_PORT = 5432
```
### Starting the server

Database migration is required during the first server initialization. In addition, re-running the migration command might be needed in a new release to apply the new schema changes (if any). It's safer to always re-run the migration script before deploying/starting a new release.

To initialize the database schema, Run Migrations with the following command:
```bash
$ compass server migrate
```

To run the Compass server use command:

```bash
$ compass server start
```

#### Using `--config` flag

```bash
$ compass server migrate --config=<path-to-file> 
```

```bash
$ compass server start --config=<path-to-file>
```

## Using the Docker 
To run the Compass server using Docker, you need to have Docker installed on your system. You can find the installation instructions [here](https://docs.docker.com/get-docker/).

You can choose to set the configuration using environment variables or a config file. The environment variables will override the config file.

If you use Docker to build compass, then configuring networking requires extra steps. Following is one of doing it by running postgres and elasticsearch inside with `docker-compose` first.

Go to the root of this project and run `docker-compose`.

```bash
$ docker-compose up
```
Once postgres and elasticsearch has been ready, we can run Compass by passing in the config of postgres and elasticsearch defined in `docker-compose.yaml` file.

### Using config file
Alternatively you can use the `compass.yaml` config file defined [above](#using-config-file) and run the following command.

```bash
$ docker run -d \
    --restart=always \
    -p 8080:8080 \
    -v $(pwd)/compass.yaml:/compass.yaml \
    --name compass-server \
    gotocompany/compass:<version> \
    server start -c /compass.yaml
```

### Using environment variables

All the configs can be passed as environment variables using underscore `_` as the delimiter between nested keys. See the example as discussed [above](#using-environment-variable)

Run the following command to start the server

```bash
$ docker run -d \
    --restart=always \
    -p 8080:8080 \
    --env-file .env \
    --name compass-server \
    gotocompany/compass:<version> \
    server start
```

## Using the Helm chart

### Pre-requisites for Helm chart
Compass can be installed in Kubernetes using the Helm chart from https://github.com/goto/charts.

Ensure that the following requirements are met:
- Kubernetes 1.14+
- Helm version 3.x is [installed](https://helm.sh/docs/intro/install/)

### Add ODPF Helm repository

Add ODPF chart repository to Helm:

```
helm repo add goto https://goto.github.io/charts/
```

You can update the chart repository by running:

```
helm repo update
```

### Setup helm values

The following table lists the configurable parameters of the Compass chart and their default values.

See full helm values guide [here](https://github.com/goto/charts/tree/main/stable/compass#values)

```yaml title="values.yaml"
app:
  image:
    repository: gotocompany/compass
    pullPolicy: Always
    tag: "0.5.0"
  container:
    command:
      - compass
    args:
      - server start
    livenessProbe:
      httpGet:
        path: /ping
        port: tcp
    readinessProbe:
      httpGet:
        path: /ping
        port: tcp

  migration:
    enabled: true
    command:
      - compass
    args:
      - server migrate

  service:
    annotations: {}

  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: contour
    hosts:
      - host: compass.example.com
        paths:
          - path: /
            pathType: ImplementationSpecific
            backend:
              service:
                # name: backend_01
                port:
                  number: 80

  config:
    COMPASS_SERVICE_PORT: 8080
    COMPASS_SERVICE_GRPC_PORT: 8081
    # COMPASS_SERVICE_HOST: 0.0.0.0
    # COMPASS_NEWRELIC_ENABLED: false
    # COMPASS_NEWRELIC_APPNAME: compass
    # COMPASS_LOG_LEVEL: info

  secretConfig: {}
    # COMPASS_ELASTICSEARCH_BROKERS: ~
    # COMPASS_ELASTICSEARCH_USERNAME: ~
    # COMPASS_ELASTICSEARCH_PASSWORD: ~
    # COMPASS_NEWRELIC_LICENSEKEY: ~
    # COMPASS_DB_HOST: ~
    # COMPASS_DB_PORT: 5432
    # COMPASS_DB_NAME: ~
    # COMPASS_DB_USER: ~
    # COMPASS_DB_PASSWORD: ~
    # COMPASS_DB_SSLMODE: disable
```

And install it with the helm command line along with the values file:

```bash
$ helm install my-release -f values.yaml goto/compass
```

## Client Initialisation

Add client configurations in the same `~/compass.yaml` file in root of current directory. Open this file to configure client. 

```yml
client:
    host: localhost:8081
    serverheaderkey_email: Compass-User-Email
    serverheadervalue_email: john.doe@example.com
```

#### Required Header/Metadata in API
Compass has a concept of [User](./concepts/user.md). In the current version, all HTTP & gRPC APIs in Compass requires an identity header/metadata in the request. The header key is configurable but the default name is `Compass-User-UUID`.

Compass APIs also expect an additional optional e-mail header. This is also configurable and the default name is `Compass-User-Email`. The purpose of having this optional e-mail header is described in the [User](./concepts/user.md) section.


If everything goes ok, you should see something like this:
```bash
time="2022-04-27T09:18:08Z" level=info msg="compass starting" version=v0.2.0
time="2022-04-27T09:18:08Z" level=info msg="connected to elasticsearch cluster" config="\"docker-cluster\" (server version 7.6.1)"
time="2022-04-27T09:18:08Z" level=info msg="New Relic monitoring is disabled."
time="2022-04-27T09:18:08Z" level=info msg="statsd metrics monitoring is disabled."
time="2022-04-27T09:18:08Z" level=info msg="connected to postgres server" host=postgres port=5432
time="2022-04-27T09:18:08Z" level=info msg="server started"
```

