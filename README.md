# DNSPulse_ConfigHub

### Выберите язык

- [English](README.md)
- [Русский](README.ru.md)

---

**DNSPulse_ConfigHub** is a central server designed for managing high-frequency monitoring agents. The main task of the server is to store configurations and a list of servers to be queried by the agents. Agents connect to the central server via the gRPC protocol to receive the necessary configuration and the list of servers that need to be queried. For successful data retrieval, an agent must provide a token and the name of the segment for authorization. All transmitted data is encrypted using the aes256 algorithm, ensuring a high level of security for information transmission.

A convenient web interface is provided for managing configurations and server lists, allowing for the uploading and editing of data in an interactive mode.


## Getting Started
To start the DNSPulse_ConfigHub server, you first need to prepare a configuration file, config.yaml. An example of the configuration file setup can be found in config-example.yaml.


## Application Build
To compile the application executable file, follow these steps:

1. Clone the repository and navigate to it in the command line:

    ```sh
    git clone git@github.com:nir0k/DNSPulse_ConfigHub.git
    cd DNSPulse_ConfigHub
    ```

2. Compile the application using the `make build` command. The compiled file will be located in the `bin` directory.


## How to Run

1. Create a configuration file in the `yaml` format. An example of the configuration file setup can be found in `config-example.yaml`.
2. Start the server using the command:

    ```sh
    ./DNSPulse_ConfigHub-linux-amd64
    ```

This project provides a powerful tool for centralized management of configurations and monitoring in distributed systems, ensuring efficient interaction between agents and the central server.