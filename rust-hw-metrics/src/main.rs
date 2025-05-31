use actix_cors::Cors;
use actix_web::{get, web, App, HttpResponse, HttpServer, Responder};
use serde::Serialize;
use std::sync::{Arc, Mutex};
use std::time::Duration;
use sysinfo::{CpuExt, System, SystemExt, DiskExt};
use tokio::time;

// Define structs for our API responses
#[derive(Serialize)]
struct CpuInfo {
    name: String,
    usage: f32,
    frequency: u64, // MHz
}

#[derive(Serialize)]
struct MemoryInfo {
    total: u64,     // KB
    used: u64,      // KB
    available: u64, // KB
    percent_used: f32,
}

#[derive(Serialize)]
struct DiskInfo {
    name: String,
    total_space: u64, // KB
    available_space: u64, // KB
    percent_used: f32,
}

#[derive(Serialize)]
struct SystemMetrics {
    hostname: String,
    uptime: u64, // seconds
    cpus: Vec<CpuInfo>,
    memory: MemoryInfo,
    disks: Vec<DiskInfo>,
}

struct AppState {
    sys: Mutex<System>,
}

// Get all metrics endpoint
#[get("/metrics")]
async fn get_metrics(data: web::Data<Arc<AppState>>) -> impl Responder {
    let mut sys = data.sys.lock().unwrap();
    
    // Refresh all system information
    sys.refresh_all();
    
    // Collect CPU information
    let cpus: Vec<CpuInfo> = sys
        .cpus()
        .iter()
        .map(|cpu| CpuInfo {
            name: cpu.name().to_string(),
            usage: cpu.cpu_usage(),
            frequency: cpu.frequency(),
        })
        .collect();
    
    // Collect memory information
    let memory = MemoryInfo {
        total: sys.total_memory(),
        used: sys.used_memory(),
        available: sys.available_memory(),
        percent_used: sys.used_memory() as f32 / sys.total_memory() as f32 * 100.0,
    };
    
    // Collect disk information
    let disks: Vec<DiskInfo> = sys
        .disks()
        .iter()
        .map(|disk| {
            let total = disk.total_space();
            let available = disk.available_space();
            let used = total - available;
            DiskInfo {
                name: disk.name().to_string_lossy().to_string(),
                total_space: total / 1024, // Convert to KB
                available_space: available / 1024, // Convert to KB
                percent_used: used as f32 / total as f32 * 100.0,
            }
        })
        .collect();
    
    // Build complete system metrics
    let metrics = SystemMetrics {
        hostname: sys.host_name().unwrap_or_else(|| "Unknown".to_string()),
        uptime: sys.uptime(),
        cpus,
        memory,
        disks,
    };
    
    HttpResponse::Ok().json(metrics)
}

// Optional: Get just CPU metrics endpoint for lighter requests
#[get("/metric/cpu")]
async fn get_cpu_metrics(data: web::Data<Arc<AppState>>) -> impl Responder {
    let mut sys = data.sys.lock().unwrap();
    
    sys.refresh_cpu();
    
    let cpus: Vec<CpuInfo> = sys
        .cpus()
        .iter()
        .map(|cpu| CpuInfo {
            name: cpu.name().to_string(),
            usage: cpu.cpu_usage(),
            frequency: cpu.frequency(),
        })
        .collect();
    
    HttpResponse::Ok().json(cpus)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize system info collector
    let mut sys = System::new_all();
    sys.refresh_all(); // Initial refresh
    
    // Create shared state
    let state = Arc::new(AppState {
        sys: Mutex::new(sys),
    });
    
    // Clone for background task
    let state_clone = state.clone();
    
    // Background task to keep CPU usage updated
    tokio::spawn(async move {
        let mut interval = time::interval(Duration::from_millis(1000));
        loop {
            interval.tick().await;
            let mut sys = state_clone.sys.lock().unwrap();
            sys.refresh_cpu();
        }
    });
    
    println!("Rust hardware metrics service starting on :8000");
    
    HttpServer::new(move || {
        let cors = Cors::permissive(); // For development - tighten this in production
        
        App::new()
            .app_data(web::Data::new(state.clone()))
            .wrap(cors)
            .service(get_metrics)
            .service(get_cpu_metrics)
    })
    .bind("127.0.0.1:8000")?
    .run()
    .await
} 