from dotenv import load_dotenv
from fastapi import FastAPI, UploadFile, File
from models import ModelSize, Languages, DeviceType
from transcribe import transcribe_file, download_models
import uvicorn
from enum import Enum

app = FastAPI()

@app.post("/transcribe/")
async def transcribe_endpoint(file: UploadFile = File(...), 
                              model_size: ModelSize = ModelSize.small, 
                              language: Languages = Languages.auto,
                              device: DeviceType = DeviceType.cpu):
    print(f"Transcribing file {file.filename} with model {model_size.value} on device {device.value}...")
    return await transcribe_file(file, model_size.value)

# Main function to run on startup

if __name__ == "__main__":
    load_dotenv()
    download_models()
    uvicorn.run(app, host="0.0.0.0", port=8000)