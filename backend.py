from http.server import BaseHTTPRequestHandler, HTTPServer
import http.client
import socket
import random
import time
import json

hostName = "0.0.0.0"
serverPort = 80
Compute = "Compute"
Storage = "Storage"
lb_host = "load-balancer:9797"  # Use the Docker service name for the load balancer

total_compute = 0
total_storage = 0

class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):
        global total_compute, total_storage  # Make sure to declare these as global to modify them

        try:
            client_hostname = socket.gethostname()
        except socket.herror:
            client_hostname = "Unknown"

        current_job_compute = self.headers.get(Compute, "0")
        current_job_storage = self.headers.get(Storage, "0")

        total_compute += int(current_job_compute)
        total_storage += int(current_job_storage)

        # Send the response
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(generate_html(client_hostname, current_job_compute, current_job_storage).encode('utf-8'))

def generate_html(client_hostname, compute, storage):
    return f"""
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Python Web Server</title>
        <style>
            body {{
                font-family: Arial, sans-serif;
                margin: 40px;
                padding: 20px;
                background-color: #f4f4f4;
            }}
            h1 {{
                color: #333;
            }}
            .container {{
                max-width: 800px;
                margin: 0 auto;
                background: #fff;
                padding: 20px;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
            }}
            p {{
                line-height: 1.6;
            }}
            .highlight {{
                color: #e74c3c;
            }}
        </style>
    </head>
    <body>
        <div class="container">
            <h1>{client_hostname} - Server Status</h1>
            <p><strong>Total Compute used:</strong> <span class="highlight">{total_compute}</span></p>
            <p><strong>Total Storage used:</strong> <span class="highlight">{total_storage}</span></p>
            <p><strong>Current Job Compute:</strong> {compute}</p>
            <p><strong>Current Job Storage:</strong> {storage}</p>
        </div>
    </body>
    </html>
    """

def http_client(host, path, body):
    # Create a connection to the server
    conn = http.client.HTTPConnection(host)
    
    try:
        # Send a GET request to the specified path with custom headers
        headers = {
            'Content-Type': 'application/json'
        }
        conn.request("GET", path, body=body, headers=headers)
        
        # Get the response from the server
        response = conn.getresponse()
        
        # Read the response content
        data = response.read()
        
        # Print the results of the request
        print("Data:", data.decode())
        
    finally:
        # Close the connection
        conn.close()    

if __name__ == "__main__":
    time.sleep(3)  # Wait for the load balancer to start
    path = "/"
    compute_vector = 20
    storage_vector = random.randint(1, 40) 

    body = json.dumps({
        "Compute-Vector": compute_vector,
        "Storage-Vector": storage_vector
    })
    http_client(lb_host, path, body)

    # Start the backend server
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print(f"Server started at http://{hostName}:{serverPort}, ratio {compute_vector / storage_vector}")
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    print("Server stopped.")
