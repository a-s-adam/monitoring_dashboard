# Monitoring Dashboard

This project is a monitoring dashboard with components written in Rust, Go, Python, and a web-based frontend. The services can be started and stopped using the provided shell scripts.

## Components

The project is structured into the following main components:

*   **`rust-hw-metrics`**: This directory contains the Rust application responsible for collecting hardware metrics from the system.
*   **`go-orchestrator`**: This directory houses the Go application that likely coordinates the operations between the different microservices (metrics collection, detection, and dashboard).
*   **`py-detector`**: This directory contains the Python application designed to perform anomaly detection on the metrics data provided by the `rust-hw-metrics` component.
*   **`web-dashboard`**: This directory holds the web application files, which serve as the user interface for displaying the collected metrics and any detected anomalies.

## Getting Started

### Installation

1.  **Clone the repository:**
    ```bash
    git clone <your-repository-url>
    cd monitoring_dashboard
    ```
    Replace `<your-repository-url>` with the actual URL of your Git repository once it's created on GitHub.

### Running the Services

1.  **Start all services:**
    To start all the components of the monitoring dashboard, execute the `start-services.sh` script from the root directory of the project:
    ```bash
    ./start-services.sh
    ```
    This script is expected to handle the initialization and execution of each service component.

2.  **Stop all services:**
    To stop all running services, execute the `stop-services.sh` script from the root directory:
    ```bash
    ./stop-services.sh
    ```

## Project Scripts

*   `start-services.sh`: A shell script to initialize and run all the necessary services for the monitoring dashboard.
*   `stop-services.sh`: A shell script to gracefully shut down all active services of the dashboard.

## Usage

[Explain how to use the dashboard, access different views, etc.]

## Contributing

[Details on how to contribute to the project, coding standards, etc.]

## License

[Specify the license for your project, e.g., MIT, Apache 2.0] 