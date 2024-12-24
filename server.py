from fastapi import FastAPI,Request
from fastapi.responses import FileResponse

app = FastAPI()

@app.get("/")
async def serve(request:Request):
    host = request.client.host
    port = request.url.port
    return FileResponse("index.html")

@app.get("/ping")
async def serve(request:Request):
    host = request.client.host
    port = request.url.port
    return host,port