from fastapi import FastAPI,Request
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
import time 



app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")

@app.get("/")
async def serve(request:Request):
    host = request.client.host
    port = request.url.port
    return host,port

@app.get("/ping")
async def serve(request:Request):
    host = request.client.host
    port = request.url.port
    return host,port