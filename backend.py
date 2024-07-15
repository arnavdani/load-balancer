from http.server import BaseHTTPRequestHandler, HTTPServer
import http.client
import socket
import random
import time

hostName = "0.0.0.0"
serverPort = 80
Compute = "Compute"
Storage = "Storage"

total_compute = 0
total_storage = 0

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        global total_compute, total_storage  # Make sure to declare these as global to modify them

        try:
            client_hostname = socket.gethostname()
        except socket.herror:
            client_hostname = "Unknown"

        compute = self.headers.get(Compute, "0")
        storage = self.headers.get(Storage, "0")

        total_compute += int(compute)
        total_storage += int(storage)

        # Send the response
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(bytes("<html><head><title>Python Web Server</title></head>", "utf-8"))
        self.wfile.write(bytes(f"<p>Backend Server Hostname: {client_hostname}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Total Compute used: {total_compute}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Total Storage used: {total_storage}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Current Job Compute: {compute}</p>", "utf-8"))
        self.wfile.write(bytes(f"<p>Current Job Storage: {storage}</p>", "utf-8"))

def http_client(host, path, headers=None):
    # Create a connection to the server
    conn = http.client.HTTPConnection(host)
    
    try:
        # Send a GET request to the specified path with custom headers
        conn.request("GET", path, headers=headers)
        
        # Get the response from the server
        response = conn.getresponse()
        
        # Read the response content
        data = response.read()
        
        # Print the status, reason, and data
        print("Status:", response.status)
        print("Reason:", response.reason)
        print("Data:", data.decode())
        
    finally:
        # Close the connection
        conn.close()    

if __name__ == "__main__":
    time.sleep(10)  # Wait for the load balancer to start
    path = "/"
    compute_vector = random.randint(1, 10) * 10
    storage_vector = random.randint(1, 10) * 10
    headers = {
        "Compute-Vector": str(compute_vector),
        "Storage-Vector": str(storage_vector)
    }
    host = "load-balancer:9797"  # Use the Docker service name for the load balancer
    http_client(host, path, headers)

    # Start the backend server
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print(f"Server started at http://{hostName}:{serverPort}")
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    print("Server stopped.")
