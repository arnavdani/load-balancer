from http.server import BaseHTTPRequestHandler, HTTPServer
import socket


hostName = "0.0.0.0"
serverPort = 80

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        # Parse the incoming request
        request_path = self.path

        # Get the client's IP address (hostname)
        client_address = self.client_address[0]
        try:
            client_hostname = socket.gethostname()
        except socket.herror:
            client_hostname = "Unknown"

        # Extract custom context (if available)

        print(self.headers)
        custom_context = self.headers.get("User-agent", "No custom context provided")

        # Send the response
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(bytes("<html><head><title>Python Web Server</title></head>", "utf-8"))
        self.wfile.write(bytes(f"<p>Request: {request_path}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Client Hostname: {client_hostname}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Custom Context: {custom_context}</p>", "utf-8"))
        self.wfile.write(bytes("<body><p>This is an enhanced example web server.</p></body></html>", "utf-8"))

if __name__ == "__main__":
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print(f"Server started at http://{hostName}:{serverPort}")
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    print("Server stopped.")
