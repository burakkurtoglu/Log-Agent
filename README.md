# Resilient Log Shipper & Agent in Go

A lightweight, high-performance, and resilient log collection agent written in Go. This project is designed to asynchronously tail multiple log files in a specific directory, stream them over a secure TCP connection to a central server, handle connection dropouts seamlessly, and adapt to OS-level log rotation events without missing a single byte.

## 🚀 Key Features

* **Asynchronous Log Tailing:** Uses Go routines to independently monitor and stream multiple `.log` files simultaneously without blocking execution.
* **Configuration Management:** Zero hardcoded values. Server addresses, log paths, and polling intervals are completely driven via an external `config.json` file.
* **Resilient Connection Handling (Retry Loop):** If the remote server goes down or is unavailable at boot, the agent implements an active retry mechanism using customizable backoff logic instead of crashing.
* **Log Rotation Awareness:** Detects file size truncation (Log Rotation) on the fly, safely resets offsets, reopens the renewed target file, and proceeds smoothly.
* **State Persistence:** Maintains read tracking states by writing atomic `.offset` files to prevent data duplication or data loss on unexpected restarts.
* **Graceful Shutdown:** Intercepts OS signals (`SIGINT` / `Ctrl+C`) via `context.Context` cancellation patterns to cleanly flush open file handles and close network sockets before exiting.
* **Ansi Color-Coded Monitoring:** Local stdout parsing classifies stream levels dynamically into `[LOG]`, `[WARNING]`, or `[CRITICAL]` outputs with appropriate console highlights.

## 🛠️ System Architecture

```text
+-------------------------------------------------------------+
|                     LIGHTHOUSE LOG AGENT                    |
|                                                             |
|  +----------------+    +----------------+    +-----------+  |
|  |  auth.log Task |    | sys.log Task   |    | app.log   |  |
|  |  (Goroutine)   |    | (Goroutine)    |    | (Gorou.)  |  |
|  +-------+--------+    +-------+--------+    +-----+-----+  |
|          |                     |                   |        |
|          +---------------------+-------------------+        |
|                                |                            |
|                                v                            |
|                    [ Shared TCP Connection ]                |
+------------------------------------------------+------------+
                                                 |
                                                 | Stream over Network
                                                 v
                                    +--------------------------+
                                    |    Central Log Server    |
                                    +--------------------------+
```
## ⚙️ Configuration (config.json)

Configure your environment metrics through the config file at the root:
```JSON
{
  "server_address": "localhost:8080",
  "log_dir": "./logs",
  "poll_interval_ms": 200,
  "retry_interval_ms": 2000
}
```

## 📦 Installation & Setup

  Clone the repository:
  ```Bash

  git clone [https://github.com/yourusername/log-agent.git](https://github.com/yourusername/log-agent.git)
  cd log-agent
   ```

  Ensure your target log directory exists and contains some .log files as specified in config.json.

  Run the log agent:
  ```Bash

  go run main.go
  ```
  🔧 Technical Deep-Dive
    Concurrency Control: Managed natively using context.Context to propagate cancellation signals deep down into isolated disk-reading workers.
    I/O Efficiency: Implements buffered reading hooks via bounded 1024-byte byte slices coupled with a non-blocking time.After multiplexed select socket structure.
