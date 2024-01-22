# aplos

## Features

🎯 Dead simple HTTP fileserver with secure defaults<br/>
🔍 Zero dependencies<br/>
🔒 Supports encryption via TLS<br/>
🩺 Health check endpoint<br/>

## Configuration

| Env Var              | Flag | Default        | Help                                                                                            |
|----------------------|------|----------------|-------------------------------------------------------------------------------------------------|
| APLOS_ADDR           | -a   | 127.0.0.1:8080 | The address to run the server on                                                                |
| APLOS_DIRECTORY      | -d   | /pub           | The directory to serve                                                                          |
| APLOS_TLS_CRT_FILE   | -c   |                | File that contains the TLS certificate                                                          |
| APLOS_TLS_KEY_FILE   | -k   |                | File that contains the TLS private key                                                          |
| APLOS_HEALTH_PATTERN | -p   | /_health       | Pattern where to expose the healthcheck handler. Set to "" to disable the health check handler. |

## Run it

```shell
docker run -v /data:/data -e APLOS_ADDR=:8080 ghcr.io/soerenschneider/aplos
```