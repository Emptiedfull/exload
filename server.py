from fastapi import FastAPI,Request,Response,WebSocket,WebSocketDisconnect
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
import time 



app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")

@app.get("/")
async def serve(request:Request):
   
    return FileResponse("experiment.html")

@app.get("/ping")
async def serve(request:Request):
    headers = {"No-Cache":"true"}
    return FileResponse("experiment.html",headers=headers)

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    await websocket.send_text("hi")
    try:
        while True:
            data = await websocket.receive_text()
            await websocket.send_text(f"Message text was: {data}")
    except WebSocketDisconnect:
        print("Client disconnected")