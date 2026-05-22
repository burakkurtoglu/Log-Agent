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
