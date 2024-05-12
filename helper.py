import requests

port = 8080
url = f"localhost:{port}/api/count?id_layanan=1"

response = requests.get(url)
json_response = response.json()
print(json_response)