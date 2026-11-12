### NetWatchPing
**NetWatchPing** is a cross-platform TCP server availability and latency monitor built in Go and Fyne
<img width="2560" height="1600" alt="NWPWork" src="https://github.com/user-attachments/assets/b41531b8-2eab-4e59-bc95-e5b25661ea58" />
* It checks servers from **targets.json** every 5 seconds and prints the ping duration
* You can add or delete a server
Automatically creates **targets.json** near the executable file

### How to Download & Run (For Users):

**Windows**: Download NetWatchPing.exe from the [Releases] page and run it.

**Linux**: Download NetWatchPing-linux from the [Releases] page, run chmod +x NetWatchPing-linux, and launch it (./NetWatchPing-linux).

### Why TCP instead of ICMP ?
* Many game servers (like Epic Games) block ICMP echo requests to mitigate DDoS, whereas TCP port checks test the actual service port.
* Unlike ICMP, which is often intercepted or spoofed locally by VPN clients, a TCP handshake travels all the way through the tunnel and reveals true routing latency and ISP/DPI timeouts.

* Built using packages ducumentation prom https://pkg.go.dev , Fyne guides from https://docs.fyne.io and some research help from Google AI
* <img width="2560" height="1600" alt="NetWatchPingWork" src="https://github.com/user-attachments/assets/840fb777-d401-4868-b708-d140a5e3505f" />
