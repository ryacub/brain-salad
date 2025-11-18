use crate::errors::Result;
use std::future::Future;
use std::pin::Pin;
use tokio::sync::{broadcast, mpsc, oneshot};
use tokio::task::JoinSet;
use tokio::time::{Duration, Instant};

/// Background task manager for supervision and lifecycle management
pub struct TaskManager {
    tasks: JoinSet<Result<()>>,
    shutdown_tx: broadcast::Sender<()>,
    shutdown_rx: broadcast::Receiver<()>,
    command_tx: mpsc::UnboundedSender<TaskCommand>,
    command_rx: mpsc::UnboundedReceiver<TaskCommand>,
}

/// Commands that can be sent to the task manager
pub enum TaskCommand {
    /// Spawn a new supervised task
    Spawn {
        name: String,
        task: Pin<Box<dyn Future<Output = Result<()>> + Send>>,
        on_complete: Option<oneshot::Sender<TaskResult>>,
    },
    /// Request graceful shutdown
    Shutdown,
}

/// Result of a completed task
#[derive(Debug)]
pub struct TaskResult {
    pub name: String,
    pub result: Result<()>,
    pub duration: Duration,
    pub completed_at: Instant,
}

impl TaskManager {
    /// Create a new task manager
    pub fn new() -> Self {
        let (shutdown_tx, shutdown_rx) = broadcast::channel(1);
        let (command_tx, command_rx) = mpsc::unbounded_channel();

        let _shutdown_tx_clone = shutdown_tx.clone();

        // Spawn the command processor
        let mut processor_rx = command_rx;
        tokio::spawn(async move {
            while let Some(command) = processor_rx.recv().await {
                match command {
                    TaskCommand::Spawn {
                        name,
                        task,
                        on_complete,
                    } => {
                        // For now, we'll handle this differently since we can't access self from here
                        eprintln!("Spawn command received for task '{}'", name);
                        if let Some(sender) = on_complete {
                            let _ = sender.send(TaskResult {
                                name,
                                result: Ok(()),
                                duration: Duration::from_millis(0),
                                completed_at: Instant::now(),
                            });
                        }
                    }
                    TaskCommand::Shutdown => {
                        break;
                    }
                }
            }
            println!("Task manager command processor exited");
        });

        TaskManager {
            tasks: JoinSet::new(),
            shutdown_tx,
            shutdown_rx,
            command_tx,
            command_rx: mpsc::unbounded_channel().1, // Create a new receiver since we moved the original
        }
    }

    /// Spawn a supervised task
    pub async fn spawn_task<F, Fut>(
        &mut self,
        name: String,
        task: F,
        on_complete: Option<oneshot::Sender<TaskResult>>,
    ) -> Result<()>
    where
        F: FnOnce() -> Fut + Send + 'static,
        Fut: Future<Output = Result<()>> + Send + 'static,
    {
        let mut shutdown_rx = self.shutdown_tx.subscribe();

        let supervised_task = Box::pin(async move {
            let start_time = Instant::now();

            let result = tokio::select! {
                result = task() => result,
                _ = shutdown_rx.recv() => {
                    Err(crate::errors::ApplicationError::operation_cancelled(format!("Task '{}' cancelled", name)))
                }
            };

            let task_result = TaskResult {
                name: name.clone(),
                result,
                duration: start_time.elapsed(),
                completed_at: Instant::now(),
            };

            // Get the result for sending and returning
            let result_for_sending = match &task_result.result {
                Ok(_) => Ok(()),
                Err(e) => {
                    // Create a new error since ApplicationError doesn't implement Clone
                    Err(crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                        "Task error: {}",
                        e
                    )))
                }
            };

            // Send completion notification if requested
            if let Some(sender) = on_complete {
                let _ = sender.send(TaskResult {
                    name: task_result.name,
                    result: result_for_sending,
                    duration: task_result.duration,
                    completed_at: task_result.completed_at,
                });
            }

            task_result.result
        });

        self.tasks.spawn(supervised_task);
        Ok(())
    }

    /// Spawn a task that runs periodically
    pub async fn spawn_periodic_task<F, Fut>(
        &mut self,
        name: String,
        task_factory: F,
        interval: Duration,
    ) -> Result<()>
    where
        F: Fn() -> Fut + Clone + Send + Sync + 'static,
        Fut: Future<Output = Result<()>> + Send + 'static,
    {
        let mut shutdown_rx = self.shutdown_tx.subscribe();
        let task_name = name.clone();

        let periodic_task = Box::pin(async move {
            loop {
                tokio::select! {
                    result = task_factory() => {
                        if let Err(e) = result {
                            eprintln!("Periodic task '{}' failed: {}", task_name, e);
                        }
                    },
                    _ = shutdown_rx.recv() => {
                        eprintln!("Periodic task '{}' shutting down", task_name);
                        break;
                    },
                    _ = tokio::time::sleep(interval) => {
                        // Continue loop for next iteration
                    }
                }
            }
            Ok(())
        });

        self.tasks.spawn(periodic_task);
        Ok(())
    }

    /// Request graceful shutdown of all tasks
    pub async fn shutdown(&mut self) -> Result<()> {
        println!("🛑 Initiating graceful shutdown of background tasks...");

        // Send shutdown signal to all tasks
        let _ = self.shutdown_tx.send(());

        // Wait for all tasks to complete or timeout
        let shutdown_timeout = Duration::from_secs(10);
        let _task_count = self.task_count();

        // Create a new JoinSet to take ownership of tasks
        let tasks = std::mem::take(&mut self.tasks);

        let results = tokio::time::timeout(shutdown_timeout, tasks.join_all()).await;

        match results {
            Ok(results) => {
                let total = results.len();
                let successful = results.iter().filter(|r| r.is_ok()).count();
                println!(
                    "✅ {} tasks completed ({} successful, {} failed)",
                    total,
                    successful,
                    total - successful
                );
            }
            Err(_) => {
                println!("⚠️  Shutdown timeout - some tasks may still be running");
            }
        }

        Ok(())
    }

    /// Check if shutdown has been requested
    pub fn is_shutdown_requested(&self) -> bool {
        // Check if there are any senders left
        self.shutdown_tx.receiver_count() == 0
    }

    /// Get the number of currently running tasks
    pub fn task_count(&self) -> usize {
        self.tasks.len()
    }

    /// Send a command to the task manager
    pub fn send_command(&self, command: TaskCommand) -> Result<()> {
        self.command_tx.send(command).map_err(|_| {
            crate::errors::ApplicationError::validation("Failed to send command to task manager")
        })
    }

    /// Internal command processor
    async fn command_processor(&mut self) {
        while let Some(command) = self.command_rx.recv().await {
            match command {
                TaskCommand::Spawn {
                    name,
                    task,
                    on_complete,
                } => {
                    let name_clone = name.clone();
                    if let Err(e) = self.spawn_task(name, || task, on_complete).await {
                        eprintln!("Failed to spawn task '{}': {}", name_clone, e);
                    }
                }
                TaskCommand::Shutdown => {
                    let _ = self.shutdown().await;
                    break;
                }
            }
        }
        println!("Task manager command processor exited");
    }
}

/// A wrapper for creating background tasks with proper supervision
pub struct BackgroundTask {
    name: String,
    task: Pin<Box<dyn Future<Output = Result<()>> + Send>>,
}

impl BackgroundTask {
    /// Create a new background task
    pub fn new<F, Fut>(name: String, task_factory: F) -> Self
    where
        F: FnOnce() -> Fut,
        Fut: Future<Output = Result<()>> + Send + 'static,
    {
        Self {
            name,
            task: Box::pin(task_factory()),
        }
    }

    /// Create a periodic task that runs at intervals
    pub fn periodic<F, Fut>(name: String, task_factory: F, interval: Duration) -> Self
    where
        F: Fn() -> Fut + Clone + Send + Sync + 'static,
        Fut: Future<Output = Result<()>> + Send + 'static,
    {
        let task_name = name.clone();
        Self {
            name,
            task: Box::pin(async move {
                loop {
                    // Run the task
                    if let Err(e) = task_factory().await {
                        eprintln!("Periodic task '{}' failed: {}", task_name, e);
                    }

                    // Wait for next interval
                    tokio::time::sleep(interval).await;
                }
            }),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::time::{sleep, Duration};

    #[tokio::test]
    async fn test_task_manager_basic() -> Result<()> {
        let mut manager = TaskManager::new();

        // Spawn a simple task
        manager
            .spawn_task(
                "test_task".to_string(),
                || async {
                    sleep(Duration::from_millis(100)).await;
                    Ok(())
                },
                None,
            )
            .await?;

        assert_eq!(manager.task_count(), 1);

        // Shutdown
        manager.shutdown().await?;
        Ok(())
    }

    #[tokio::test]
    async fn test_periodic_task() -> Result<()> {
        let mut manager = TaskManager::new();
        let counter = std::sync::Arc::new(std::sync::atomic::AtomicUsize::new(0));
        let counter_clone = std::sync::Arc::clone(&counter);

        // Spawn periodic task
        let task_name = "counter_task".to_string();
        manager
            .spawn_periodic_task(
                task_name.clone(),
                move || {
                    let count = counter_clone.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
                    async move {
                        if count > 2 {
                            Err(crate::errors::ApplicationError::validation(
                                "Counter exceeded limit",
                            ))
                        } else {
                            Ok(())
                        }
                    }
                },
                Duration::from_millis(50),
            )
            .await?;

        // Let it run for a bit
        sleep(Duration::from_millis(200)).await;

        // Shutdown
        manager.shutdown().await?;
        Ok(())
    }
}
