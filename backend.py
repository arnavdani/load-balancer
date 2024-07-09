from http.server import BaseHTTPRequestHandler, HTTPServer
import socket


hostName = "0.0.0.0"
serverPort = 80

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        ## BOTH OF THESE NOT USED NOW, WILL BE LATER
        # path of request
        # request_path = self.path
        # client IP
        # client_address = self.client_address[0]

        try:
            client_hostname = socket.gethostname()
        except socket.herror:
            client_hostname = "Unknown"

        # Extract custom context (if available)

        # print(self.headers)
        custom_context = self.headers.get("JobSize", "No custom context provided")

        # Send the response
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(bytes("<html><head><title>Python Web Server</title></head>", "utf-8"))
        self.wfile.write(bytes(f"<p>Backend Server Hostname: {client_hostname}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>JobSize: {custom_context}</p>", "utf-8"))

if __name__ == "__main__":
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print(f"Server started at http://{hostName}:{serverPort}")
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    print("Server stopped.")
