from faster_whisper import WhisperModel, download_model
from enum import Enum
from models import DeviceType
import json
import os

# TODO: Do not download models every time, only if they are not present
def download_models():
    # Get model list (comma separated) from environment variable
    model_list = os.environ.get("WHISPER_MODELS", "tiny,base,small")
    model_list = model_list.split(",")
    for model in model_list:
        download_model(model)

def segment_to_dict(segment):
    segment = segment._asdict()
    if segment["words"] is not None:
        segment["words"] = [word._asdict() for word in segment["words"]]
    return segment

async def transcribe_file(file, 
                          model_size, 
                          language: str | None = None, 
                          device: DeviceType = DeviceType.cpu):
    if language == "auto":
        language = None
    model = WhisperModel(model_size, device, compute_type="int8")

    # Save the uploaded file temporarily
    with open(file.filename, 'wb') as buffer:
        contents = await file.read()  # async read
        buffer.write(contents)

    # Transcribe the file
    segments, info = model.transcribe(file.filename, beam_size=5, word_timestamps=True, language=language)

    # Delete the temporary file
    os.remove(file.filename)

    segments = [segment_to_dict(segment) for segment in segments]
    text = " ".join([segment["text"] for segment in segments])
    
    # Add language to segments
    segments = {"language": info.language, "text": text, "language_confidence": info.language_probability, "duration": info.duration, "segments": segments}

    print("Transcription complete.")
    return segments