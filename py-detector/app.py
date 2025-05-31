from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import numpy as np
import logging
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger("anomaly-detector")

app = FastAPI(title="Hardware Metrics Anomaly Detector")

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class AnomalyRequest(BaseModel):
    cpu_values: List[float]
    memory_values: List[float]

class AnomalyResponse(BaseModel):
    cpu_anomaly: bool
    memory_anomaly: bool
    cpu_score: float = None
    memory_score: float = None

def detect_anomaly(values, threshold=3.0):
    """
    Detect anomalies using Z-score method
    Returns (is_anomaly, score)
    """
    if len(values) < 5:
        return False, 0.0
        
    values_array = np.array(values)
    
    # Get the latest value
    latest = values_array[-1]
    
    # Calculate historical mean and std using all but the latest value
    history = values_array[:-1]
    mean = np.mean(history)
    std = np.std(history)
    
    # Avoid division by zero
    if std == 0:
        std = 0.1  # Small non-zero value
    
    # Calculate Z-score for latest value
    z_score = abs((latest - mean) / std)
    
    # If Z-score is beyond threshold, it's an anomaly
    is_anomaly = z_score > threshold
    
    return is_anomaly, float(z_score)

@app.post("/detect", response_model=AnomalyResponse)
async def detect_anomalies(request: AnomalyRequest):
    if not request.cpu_values or not request.memory_values:
        raise HTTPException(status_code=400, detail="CPU and memory values are required")
    
    # Get anomaly detection results
    cpu_anomaly, cpu_score = detect_anomaly(request.cpu_values)
    memory_anomaly, memory_score = detect_anomaly(request.memory_values)
    
    if cpu_anomaly or memory_anomaly:
        logger.info(f"Anomaly detected - CPU: {cpu_anomaly} ({cpu_score:.2f}), "
                   f"Memory: {memory_anomaly} ({memory_score:.2f})")
    
    return AnomalyResponse(
        cpu_anomaly=cpu_anomaly,
        memory_anomaly=memory_anomaly,
        cpu_score=cpu_score,
        memory_score=memory_score
    )

if __name__ == "__main__":
    logger.info("Starting anomaly detection service on port 8001...")
    uvicorn.run(app, host="127.0.0.1", port=8001) 